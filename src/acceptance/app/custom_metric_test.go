package app_test

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
		initialInstanceCount int
		policy               string
		doneChan             chan bool
		doneAcceptChan       chan bool
		ticker               *time.Ticker
	)

	Context("when scaling by custom metrics", func() {

		JustBeforeEach(func() {
			appName = generator.PrefixedRandomName("autoscaler", "nodeapp")
			countStr := strconv.Itoa(initialInstanceCount)
			createApp := cf.Cf("push", appName, "--no-start", "--no-route", "-i", countStr, "-b", cfg.NodejsBuildpackName, "-m", "128M", "-p", config.NODE_APP).Wait(cfg.CfPushTimeoutDuration())
			Expect(createApp).To(Exit(0), "failed creating app")

			mapRouteToApp := cf.Cf("map-route", appName, cfg.AppsDomain, "--hostname", appName).Wait(cfg.DefaultTimeoutDuration())
			Expect(mapRouteToApp).To(Exit(0), "failed to map route to app")

			guid := cf.Cf("app", appName, "--guid").Wait(cfg.DefaultTimeoutDuration())
			Expect(guid).To(Exit(0))
			appGUID = strings.TrimSpace(string(guid.Out.Contents()))
			CreatePolicy(appName, appGUID, policy)
			if !cfg.IsServiceOfferingEnabled() {
				CreateCustomMetricCred(appName, appGUID)
			}
			Expect(cf.Cf("start", appName).Wait(cfg.CfPushTimeoutDuration())).To(Exit(0))
			WaitForNInstancesRunning(appGUID, initialInstanceCount, cfg.DefaultTimeoutDuration())

			doneChan = make(chan bool)
			doneAcceptChan = make(chan bool)
		})

		AfterEach(func() {
			close(doneChan)
			Eventually(doneAcceptChan, 30*time.Second).Should(Receive())

			DeletePolicy(appName, appGUID)
			if !cfg.IsServiceOfferingEnabled() {
				DeleteCustomMetricCred(appGUID)
			}

			Expect(cf.Cf("delete", appName, "-f", "-r").Wait(cfg.DefaultTimeoutDuration())).To(Exit(0))

		})

		Context("when test_metric is less than scaling in threshold", func() {

			BeforeEach(func() {
				policy = GenerateDynamicScaleInPolicy(cfg, 1, 2, "test_metric", 500)
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
								return helpers.CurlApp(cfg, appName, "/custom-metrics/test_metric/100")
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

		Context("when test_metric is more than scaling in threshold", func() {

			BeforeEach(func() {
				policy = GenerateDynamicScaleOutPolicy(cfg, 1, 2, "test_metric", 500)
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
								return helpers.CurlApp(cfg, appName, "/custom-metrics/test_metric/800")
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
