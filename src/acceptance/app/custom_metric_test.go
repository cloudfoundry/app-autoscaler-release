package app_test

import (
	"acceptance"
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
		CreateCustomMetricCred(cfg, appName, appGUID)
		StartApp(appName, cfg.CfPushTimeoutDuration())
	})
	AfterEach(AppAfterEach)

	Context("when scaling by custom metrics", func() {
		It("should scale out and scale in", Label(acceptance.LabelSmokeTests), func() {
			By("Scale out to 2 instances")
			scaleOut := func() (int, error) {
				SendMetric(cfg, appName, 550)
				return RunningInstances(appGUID, 5*time.Second)
			}
			Eventually(scaleOut, 5*time.Minute, 15*time.Second).Should(Equal(2))

			By("Scale in to 1 instances")
			scaleIn := func() (int, error) {
				SendMetric(cfg, appName, 100)
				return RunningInstances(appGUID, 5*time.Second)
			}
			Eventually(scaleIn, 5*time.Minute, 15*time.Second).Should(Equal(1))
		})
	})

	Context("when adding custom-metrics via mtls", func() {
		It("should successfully add a metric using the app", func() {
			By("adding policy so test_metric is allowed")
			policy = GenerateDynamicScaleOutAndInPolicy(1, 2, "test_metric", 500, 500)
			By("sending metric via mtls endpoint")
			SendMetricMTLS(cfg, appName, 10)
			GinkgoWriter.Println("")
		})
	})

	Context("when scaling by custom metrics via mtls", func() {
		It("should scale out and scale in", Label(acceptance.LabelSmokeTests), func() {
			By("Scale out to 2 instances")
			scaleOut := func() (int, error) {
				SendMetricMTLS(cfg, appName, 550)
				return RunningInstances(appGUID, 5*time.Second)
			}
			Eventually(scaleOut, 5*time.Minute, 15*time.Second).Should(Equal(2))

			By("Scale in to 1 instance")
			scaleIn := func() (int, error) {
				SendMetricMTLS(cfg, appName, 100)
				return RunningInstances(appGUID, 5*time.Second)
			}
			Eventually(scaleIn, 5*time.Minute, 15*time.Second).Should(Equal(1))
		})
	})
})
