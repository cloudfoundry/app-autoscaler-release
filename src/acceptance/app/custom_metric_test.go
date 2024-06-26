package app_test

import (
	"acceptance"
	"acceptance/config"
	. "acceptance/helpers"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AutoScaler custom metrics policy", func() {
	var (
		policy string
		err    error
	)
	BeforeEach(func() {
		policy = GenerateDynamicScaleOutAndInPolicy(1, 2, "test_metric", 500, 500)
		appName = CreateTestApp(cfg, "node-custom-metric", 1)
		appGUID, err = GetAppGuid(cfg, appName)
		Expect(err).NotTo(HaveOccurred())
		instanceName = CreatePolicy(cfg, appName, appGUID, policy)
		StartApp(appName, cfg.CfPushTimeoutDuration())
	})
	AfterEach(AppAfterEach)

	// This test will fail if credential-type is set to X509 int autoscaler broker. Therefore, we will only support mtls
	// connection with custom metrics apis.
	Context("when scaling by custom metrics", func() {
		It("should scale out and scale in", Label(acceptance.LabelSmokeTests), func() {
			By("Scale out to 2 instances")
			scaleOut := sendMetricToAutoscaler(cfg, appGUID, appName, 550, false)
			Eventually(scaleOut).
				WithTimeout(5 * time.Minute).
				WithPolling(15 * time.Second).
				Should(Equal(2))

			By("Scale in to 1 instances")
			scaleIn := sendMetricToAutoscaler(cfg, appGUID, appName, 100, false)
			Eventually(scaleIn).
				WithTimeout(5 * time.Minute).
				WithPolling(15 * time.Second).
				Should(Equal(1))

		})
	})

	Context("when scaling by custom metrics via mtls", func() {
		It("should scale out and scale in", Label(acceptance.LabelSmokeTests), func() {
			By("Scale out to 2 instances")
			scaleOut := sendMetricToAutoscaler(cfg, appGUID, appName, 550, true)
			Eventually(scaleOut).
				WithTimeout(5 * time.Minute).
				WithPolling(15 * time.Second).
				Should(Equal(2))

			By("Scale in to 1 instance")
			scaleIn := sendMetricToAutoscaler(cfg, appGUID, appName, 100, true)
			Eventually(scaleIn).
				WithTimeout(5 * time.Minute).
				WithPolling(15 * time.Second).
				Should(Equal(1))

		})
	})
})

func sendMetricToAutoscaler(config *config.Config, appGUID string, appName string, metricThreshold int, mtls bool) func() (int, error) {
	return func() (int, error) {
		if mtls {
			SendMetricMTLS(config, appName, metricThreshold)
		} else {
			SendMetric(config, appName, metricThreshold)
		}
		return RunningInstances(appGUID, 5*time.Second)
	}
}
