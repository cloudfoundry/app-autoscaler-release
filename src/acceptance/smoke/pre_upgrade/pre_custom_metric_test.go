package pre_upgrade_test

import (
	"acceptance/app"
	"acceptance/helpers"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AutoScaler custom metrics policy", func() {
	var (
		appName              string
		appGUID              string
		initialInstanceCount int
		policy               string
	)

	BeforeEach(func() {
		initialInstanceCount = 1
		policy = helpers.GenerateDynamicScaleOutAndInPolicy(1, 2, "test_metric", 500, 500)
		appName = CreateApp("node-custom-metric", initialInstanceCount)
		appGUID = helpers.GetAppGuid(cfg, appName)
		_ = helpers.CreatePolicy(cfg, appName, appGUID, policy)
		if !cfg.IsServiceOfferingEnabled() {
			CreateCustomMetricCred(appName, appGUID)
		}
		StartApp(appName)
		helpers.WaitForNInstancesRunning(appGUID, initialInstanceCount, cfg.DefaultTimeoutDuration())
	})

	Context("when scaling by custom metrics", func() {
		It("should scale out and scale in", func() {
			By("Scale out to 2 instances")
			scaleOut := func() int {
				SendMetric(appName, 550)
				return helpers.RunningInstances(appGUID, 5*time.Second)
			}
			Eventually(scaleOut, 5*time.Minute, 15*time.Second).Should(Equal(2))

			By("Scale in to 1 instances")
			scaleIn := func() int {
				SendMetric(appName, 100)
				return helpers.RunningInstances(appGUID, 5*time.Second)
			}
			Eventually(scaleIn, 5*time.Minute, 15*time.Second).Should(Equal(1))
		})
	})
})

func SendMetric(appName string, metric int) {
	curler := app.NewAppCurler(cfg)
	Eventually(func() string {
		response := curler.Curl(appName, fmt.Sprintf("/custom-metrics/test_metric/%d", metric), 60*time.Second)
		if response == "" {
			return "success"
		}
		return response
	}, cfg.DefaultTimeoutDuration(), 5*time.Second).Should(ContainSubstring("success"))
}
