package app_test

import (
	"acceptance/helpers"
	"fmt"
	"os"
	"time"

	cfh "github.com/KevinJCross/cf-test-helpers/v2/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AutoScaler custom metrics policy", func() {
	var (
		appName string
		appGUID string
		policy  string
	)
	BeforeEach(func() {
		policy = helpers.GenerateDynamicScaleOutAndInPolicy(1, 2, "test_metric", 500, 500)
		appName = helpers.CreateTestApp(cfg, "node-custom-metric", 1)
		appGUID = helpers.GetAppGuid(cfg, appName)
		instanceName = helpers.CreatePolicy(cfg, appName, appGUID, policy)
		if !cfg.IsServiceOfferingEnabled() {
			helpers.CreateCustomMetricCred(cfg, appName, appGUID)
		}
		helpers.StartApp(appName, cfg.CfPushTimeoutDuration())
		helpers.WaitForNInstancesRunning(appGUID, 1, cfg.DefaultTimeoutDuration())
	})

	AfterEach(func() {
		if os.Getenv("SKIP_TEARDOWN") == "true" {
			fmt.Println("Skipping Teardown...")
		} else {
			DeletePolicy(appName, appGUID)
			if !cfg.IsServiceOfferingEnabled() {
				helpers.DeleteCustomMetricCred(cfg, appGUID)
			}
			helpers.DeleteTestApp(appName, cfg.DefaultTimeoutDuration())
		}
	})

	Context("when scaling by custom metrics", func() {
		It("should scale out and scale in", func() {
			By("Scale out to 2 instances")
			scaleOut := func() int {
				helpers.SendMetric(cfg, appName, 550)
				return helpers.RunningInstances(appGUID, 5*time.Second)
			}
			Eventually(scaleOut, 5*time.Minute, 15*time.Second).Should(Equal(2))

			By("Scale in to 1 instances")
			scaleIn := func() int {
				helpers.SendMetric(cfg, appName, 100)
				return helpers.RunningInstances(appGUID, 5*time.Second)
			}
			Eventually(scaleIn, 5*time.Minute, 15*time.Second).Should(Equal(1))
		})
	})

	Context("when adding custom-metrics via mtls", func() {
		It("should successfully add a metric using the app", func() {
			By("adding policy so test_metric is allowed")
			policy = helpers.GenerateDynamicScaleOutAndInPolicy(1, 2, "test_metric", 500, 500)
			By("sending metric via mtls endpoint")
			cfh.CurlAppWithTimeout(cfg, appName, "/custom-metrics/mtls/test_metric/10", 10*time.Second, "-f")
		})
	})
})
