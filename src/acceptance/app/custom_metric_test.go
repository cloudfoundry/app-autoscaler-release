package app_test

import (
	"acceptance"
	"acceptance/helpers"
	"fmt"
	"os"
	"time"

	"github.com/KevinJCross/cf-test-helpers/v2/cf"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("AutoScaler custom metrics policy", func() {
	var (
		appGUID string
		policy  string
	)
	BeforeEach(func() {
		policy = helpers.GenerateDynamicScaleOutAndInPolicy(1, 2, "test_metric", 500, 500)
		appName = helpers.CreateTestApp(cfg, "node-custom-metric", 1)
		appGUID = helpers.GetAppGuid(cfg, appName)
		instanceName = helpers.CreatePolicy(cfg, appName, appGUID, policy)
		helpers.CreateCustomMetricCred(cfg, appName, appGUID)
		helpers.StartApp(appName, cfg.CfPushTimeoutDuration())
	})

	AfterEach(func() {
		if os.Getenv("SKIP_TEARDOWN") == "true" {
			fmt.Println("Skipping Teardown...")
		} else {
			Eventually(cf.Cf("logs", appName, "--recent"), cfg.DefaultTimeoutDuration()).Should(Exit())
			DeletePolicy(appName, appGUID)
			if !cfg.IsServiceOfferingEnabled() {
				helpers.DeleteCustomMetricCred(cfg, appGUID)
			}
			helpers.DeleteTestApp(appName, cfg.DefaultTimeoutDuration())
		}
	})

	Context("when scaling by custom metrics", func() {
		It("should scale out and scale in", Label(acceptance.LabelSmokeTests), func() {
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
			cf.Cf("curl", "/custom-metrics/mtls/test_metric/10", "-f").Wait(cfg.DefaultTimeoutDuration())
		})
	})
})
