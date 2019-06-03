package app

import (
	"acceptance/config"
	. "acceptance/helpers"

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

var _ = Describe("AutoScaler custom metrics policy", func() {
	var (
		appName              string
		appGUID              string
		instanceName         string
		initialInstanceCount int
		policy               string
		doneChan             chan bool
		doneAcceptChan       chan bool
		ticker               *time.Ticker
	)

	Context("when scaling by custom metrics", func() {

		JustBeforeEach(func() {
			instanceName = generator.PrefixedRandomName("autoscaler", "service")
			createService := cf.Cf("create-service", cfg.ServiceName, cfg.ServicePlan, instanceName).Wait(cfg.DefaultTimeoutDuration())
			Expect(createService).To(Exit(0), "failed creating service")

			appName = generator.PrefixedRandomName("autoscaler", "nodeapp")
			countStr := strconv.Itoa(initialInstanceCount)
			createApp := cf.Cf("push", appName, "--no-start", "-i", countStr, "-b", cfg.NodejsBuildpackName, "-m", "128M", "-p", config.NODE_APP, "-d", cfg.AppsDomain).Wait(cfg.CfPushTimeoutDuration())
			Expect(createApp).To(Exit(0), "failed creating app")

			guid := cf.Cf("app", appName, "--guid").Wait(cfg.DefaultTimeoutDuration())
			Expect(guid).To(Exit(0))
			appGUID = strings.TrimSpace(string(guid.Out.Contents()))

			bindService := cf.Cf("bind-service", appName, instanceName, "-c", policy).Wait(cfg.DefaultTimeoutDuration())
			Expect(bindService).To(Exit(0), "failed binding service to app with a policy ")

			Expect(cf.Cf("start", appName).Wait(cfg.CfPushTimeoutDuration())).To(Exit(0))
			WaitForNInstancesRunning(appGUID, initialInstanceCount, cfg.DetectTimeoutDuration())

			doneChan = make(chan bool)
			doneAcceptChan = make(chan bool)
		})

		AfterEach(func() {
			close(doneChan)
			Eventually(doneAcceptChan, 30*time.Second).Should(Receive())
			unbindService := cf.Cf("unbind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
			Expect(unbindService).To(Exit(0), "failed unbinding service from app")

			Expect(cf.Cf("delete", appName, "-f", "-r").Wait(cfg.DefaultTimeoutDuration())).To(Exit(0))
			deleteService := cf.Cf("delete-service", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
			Expect(deleteService).To(Exit(0))
		})

		Context("when test-metric is less than scaling in threshold", func() {

			BeforeEach(func() {
				policy = GenerateDynamicScaleInPolicy(cfg, 1, 2, "test-metric", 500)
				initialInstanceCount = 2
			})

			JustBeforeEach(func() {
				ticker = time.NewTicker(15 * time.Second)
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
								return helpers.CurlApp(cfg, appName, "/custom-metrics/test-metric/100")
							}, cfg.DefaultTimeoutDuration(), 5*time.Second).Should(ContainSubstring("success"))
						}
					}
				}(doneChan)
			})

			It("should scale in", func() {
				finishTime := time.Duration(interval*2)*time.Second + 5*time.Minute
				WaitForNInstancesRunning(appGUID, 1, finishTime)
			})
		})

		Context("when test-metric is more than scaling in threshold", func() {

			BeforeEach(func() {
				policy = GenerateDynamicScaleInPolicy(cfg, 1, 2, "test-metric", 500)
				initialInstanceCount = 1
			})

			JustBeforeEach(func() {
				ticker = time.NewTicker(15 * time.Second)
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
								return helpers.CurlApp(cfg, appName, "/custom-metrics/test-metric/800")
							}, cfg.DefaultTimeoutDuration(), 5*time.Second).Should(ContainSubstring("success"))
						}
					}
				}(doneChan)
			})

			It("should scale out", func() {
				finishTime := time.Duration(interval*2)*time.Second + 5*time.Minute
				WaitForNInstancesRunning(appGUID, 2, finishTime)
			})
		})

	})

})
