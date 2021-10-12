package app_test

import (
	"acceptance/helpers"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
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

	Context("when scaling by custom metrics via MTLS", func() {
		It("should scale out and scale in", func() {
			By("Scale out to 2 instances")
			scaleOut := func() int {
				helpers.SendMtlsMetric(cfg, appName, 550)
				return helpers.RunningInstances(appGUID, 5*time.Second)
			}
			Eventually(scaleOut, 5*time.Minute, 15*time.Second).Should(Equal(2))

			By("Scale in to 1 instances")
			scaleIn := func() int {
				helpers.SendMtlsMetric(cfg, appName, 100)
				return helpers.RunningInstances(appGUID, 5*time.Second)
			}
			Eventually(scaleIn, 5*time.Minute, 15*time.Second).Should(Equal(1))
		})
	})
})
