package app

import (
	"acceptance/config"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"strconv"
	"strings"
	"time"
)

var _ = Describe("AutoScaler recurring schedule policy", func() {
	var (
		appName              string
		appGUID              string
		instanceName         string
		initialInstanceCount int
		daysOfMonthOrWeek    Days
		startTime            time.Time
		endTime              time.Time
		doneChan             chan bool
		doneAcceptChan       chan bool
		ticker               *time.Ticker
	)

	BeforeEach(func() {
		instanceName = generator.PrefixedRandomName("autoscaler", "service")
		createService := cf.Cf("create-service", cfg.ServiceName, cfg.ServicePlan, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(createService).To(Exit(0), "failed creating service")

		initialInstanceCount = 1
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

	Context("when scaling by recurring schedule", func() {

		JustBeforeEach(func() {

			Expect(cf.Cf("start", appName).Wait(cfg.CfPushTimeoutDuration())).To(Exit(0))
			waitForNInstancesRunning(appGUID, initialInstanceCount, cfg.DetectTimeoutDuration())

			location, err := time.LoadLocation("GMT")
			Expect(err).NotTo(HaveOccurred())
			startTime, endTime = getStartAndEndTime(location, 70*time.Second, time.Duration(interval+360)*time.Second)
			policyStr := generateDynamicAndRecurringSchedulePolicy(1, 4, 4, "GMT", startTime, endTime, daysOfMonthOrWeek, 2, 5, 3)

			bindService := cf.Cf("bind-service", appName, instanceName, "-c", policyStr).Wait(cfg.DefaultTimeoutDuration())
			Expect(bindService).To(Exit(0), "failed binding service to app with a policy ")
			doneChan = make(chan bool)
			doneAcceptChan = make(chan bool)
		})

		AfterEach(func() {
			close(doneChan)
			Eventually(doneAcceptChan, 60*time.Second).Should(Receive())
			unbindService := cf.Cf("unbind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
			Expect(unbindService).To(Exit(0), "failed unbinding service from app")

		})

		Context("with days of month", func() {
			BeforeEach(func() {
				daysOfMonthOrWeek = daysOfMonth
			})

			JustBeforeEach(func() {
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

			It("should scale", func() {
				By("setting to initial_min_instance_count")
				jobRunTime := startTime.Add(4 * time.Minute).Sub(time.Now())
				waitForNInstancesRunning(appGUID, 3, jobRunTime)

				By("setting schedule's instance_min_count")
				jobRunTime = endTime.Sub(time.Now())
				Eventually(func() int {
					return runningInstances(appGUID, jobRunTime)
				}, jobRunTime, 15*time.Second).Should(Equal(2))

				jobRunTime = endTime.Sub(time.Now())
				Consistently(func() int {
					return runningInstances(appGUID, jobRunTime)
				}, jobRunTime, 15*time.Second).Should(Equal(2))

				By("setting to default instance_min_count")
				waitForNInstancesRunning(appGUID, 1, time.Duration(interval+60)*time.Second+5*time.Minute)
			})

		})

		Context("with days of week", func() {
			BeforeEach(func() {
				daysOfMonthOrWeek = daysOfWeek
			})

			JustBeforeEach(func() {
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

			It("should scale", func() {
				By("setting to initial_min_instance_count")
				jobRunTime := startTime.Add(3 * time.Minute).Sub(time.Now())
				waitForNInstancesRunning(appGUID, 3, jobRunTime)

				By("setting schedule's instance_min_count")
				jobRunTime = endTime.Sub(time.Now())
				Eventually(func() int {
					return runningInstances(appGUID, jobRunTime)
				}, jobRunTime, 15*time.Second).Should(Equal(2))

				jobRunTime = endTime.Sub(time.Now())
				Consistently(func() int {
					return runningInstances(appGUID, jobRunTime)
				}, jobRunTime, 15*time.Second).Should(Equal(2))

				By("setting to default instance_min_count")
				waitForNInstancesRunning(appGUID, 1, time.Duration(interval+60)*time.Second+5*time.Minute)
			})
		})
	})

})
