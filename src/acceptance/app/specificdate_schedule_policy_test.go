package app

import (
	"acceptance/config"
	. "acceptance/helpers"

	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("AutoScaler specific date schedule policy", func() {
	var (
		appName              string
		appGUID              string
		instanceName         string
		initialInstanceCount int
		location             *time.Location
		startDateTime        time.Time
		endDateTime          time.Time
		doneChan             chan bool
		doneAcceptChan       chan bool
		ticker               *time.Ticker
	)

	BeforeEach(func() {
		initialInstanceCount = 1
		instanceName = generator.PrefixedRandomName("autoscaler", "service")
		createService := cf.Cf("create-service", cfg.ServiceName, cfg.ServicePlan, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(createService).To(Exit(0), "failed creating service")
	})

	JustBeforeEach(func() {
		appName = generator.PrefixedRandomName("autoscaler", "nodeapp")
		countStr := strconv.Itoa(initialInstanceCount)
		createApp := cf.Cf("push", appName, "--no-start", "-i", countStr, "-b", cfg.NodejsBuildpackName, "-m", "128M", "-p", config.NODE_APP, "-d", cfg.AppsDomain).Wait(cfg.CfPushTimeoutDuration())
		Expect(createApp).To(Exit(0), "failed creating app")

		guid := cf.Cf("app", appName, "--guid").Wait(cfg.DefaultTimeoutDuration())
		Expect(guid).To(Exit(0))
		appGUID = strings.TrimSpace(string(guid.Out.Contents()))
	})

	AfterEach(func() {
		Expect(cf.Cf("delete", appName, "-f", "-r").Wait(cfg.DefaultTimeoutDuration())).To(Exit(0))

		deleteService := cf.Cf("delete-service", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
		Expect(deleteService).To(Exit(0))
	})

	Context("when scaling by specific date schedule ", func() {

		JustBeforeEach(func() {

			Expect(cf.Cf("start", appName).Wait(cfg.CfPushTimeoutDuration())).To(Exit(0))
			WaitForNInstancesRunning(appGUID, initialInstanceCount, cfg.DetectTimeoutDuration())

			location, _ = time.LoadLocation("GMT")
			timeNowInTimeZoneWithOffset := time.Now().In(location).Add(70 * time.Second).Truncate(time.Minute)
			startDateTime = timeNowInTimeZoneWithOffset
			endDateTime = timeNowInTimeZoneWithOffset.Add(time.Duration(interval+360) * time.Second)

			policyStr := GenerateDynamicAndSpecificDateSchedulePolicy(cfg, 1, 4, 80, "GMT", startDateTime, endDateTime, 2, 5, 3)
			bindService := cf.Cf("bind-service", appName, instanceName, "-c", policyStr).Wait(cfg.DefaultTimeoutDuration())
			Expect(bindService).To(Exit(0), "failed binding service to app with a policy ")

			doneChan = make(chan bool)
			doneAcceptChan = make(chan bool)
			ticker = time.NewTicker(30 * time.Second)
			go func(chan bool) {
				defer GinkgoRecover()
				for {
					select {
					case <-doneChan:
						ticker.Stop()
						doneAcceptChan <- true
						return
					case <-ticker.C:
						Eventually(func() string {
							return helpers.CurlApp(cfg, appName, "/fast")
						}, cfg.DefaultTimeoutDuration(), 1*time.Second).Should(ContainSubstring("dummy application with fast response"))
					}
				}
			}(doneChan)
		})

		AfterEach(func() {
			close(doneChan)
			Eventually(doneAcceptChan, 60*time.Second).Should(Receive())
			unbindService := cf.Cf("unbind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
			Expect(unbindService).To(Exit(0), "failed unbinding service from app")
		})

		It("should scale", func() {
			By("setting to initial_min_instance_count")
			jobRunTime := startDateTime.Add(4 * time.Minute).Sub(time.Now())
			WaitForNInstancesRunning(appGUID, 3, jobRunTime)

			By("setting to schedule's instance_min_count")
			jobRunTime = endDateTime.Sub(time.Now())
			Eventually(func() int {
				return RunningInstances(appGUID, jobRunTime)
			}, jobRunTime, 15*time.Second).Should(Equal(2))

			jobRunTime = endDateTime.Sub(time.Now())
			Consistently(func() int {
				return RunningInstances(appGUID, jobRunTime)
			}, jobRunTime, 15*time.Second).Should(Equal(2))

			By("setting to default instance_min_count")
			WaitForNInstancesRunning(appGUID, 1, time.Duration(interval+60)*time.Second+5*time.Minute)

		})

	})

})
