package app_test

import (
	"acceptance"
	"acceptance/config"
	. "acceptance/helpers"
	"fmt"
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
		appToScaleName = CreateTestApp(cfg, "node-custom-metric", 1)
		appToScaleGUID, err = GetAppGuid(cfg, appToScaleName)
		Expect(err).NotTo(HaveOccurred())
		instanceName = CreatePolicy(cfg, appToScaleName, appToScaleGUID, policy)
		StartApp(appToScaleName, cfg.CfPushTimeoutDuration())
	})
	AfterEach(AppAfterEach)

	// This test will fail if credential-type is set to X509 in autoscaler broker.
	// Therefore, only mtls connection will be supported for custom metrics in future
	Context("when scaling by custom metrics", func() {
		It("should scale out and scale in", Label(acceptance.LabelSmokeTests), func() {
			By("Scale out to 2 instances")
			scaleOut := sendMetricToAutoscaler(cfg, appToScaleGUID, appToScaleName, 550, false)
			Eventually(scaleOut).
				WithTimeout(5 * time.Minute).
				WithPolling(15 * time.Second).
				Should(Equal(2))

			By("Scale in to 1 instances")
			scaleIn := sendMetricToAutoscaler(cfg, appToScaleGUID, appToScaleName, 100, false)
			Eventually(scaleIn).
				WithTimeout(5 * time.Minute).
				WithPolling(15 * time.Second).
				Should(Equal(1))

		})
	})

	Context("when scaling by custom metrics via mtls", func() {
		It("should scale out and scale in", Label(acceptance.LabelSmokeTests), func() {
			By("Scale out to 2 instances")
			scaleOut := sendMetricToAutoscaler(cfg, appToScaleGUID, appToScaleName, 550, true)
			Eventually(scaleOut).
				WithTimeout(5 * time.Minute).
				WithPolling(15 * time.Second).
				Should(Equal(2))

			By("Scale in to 1 instance")
			scaleIn := sendMetricToAutoscaler(cfg, appToScaleGUID, appToScaleName, 100, true)
			Eventually(scaleIn).
				WithTimeout(5 * time.Minute).
				WithPolling(15 * time.Second).
				Should(Equal(1))

		})
	})
	Describe("Custom metrics policy with neighbour app", func() {
		JustBeforeEach(func() {
			neighbourAppName = CreateTestApp(cfg, "go-neighbour-app", 1)
			neighbourAppGUID, err = GetAppGuid(cfg, neighbourAppName)
			Expect(err).NotTo(HaveOccurred())
			err := BindServiceToAppWithPolicy(cfg, neighbourAppName, instanceName, policy)
			Expect(err).NotTo(HaveOccurred())
			StartApp(neighbourAppName, cfg.CfPushTimeoutDuration())
		})
		Context("neighbour app send custom metrics for app B via mtls", func() {
			BeforeEach(func() {
				policy = GenerateBindingsWithScalingPolicy("bound_app", 1, 2, "test_metric", 500, 100)
			})
			It("should scale out and scale in app B", Label(acceptance.LabelSmokeTests), func() {
				By(fmt.Sprintf("Scale out %s to 2 instance", appToScaleName))
				scaleOut := sendMetricToAutoscaler(cfg, appToScaleGUID, neighbourAppName, 550, true)
				Eventually(scaleOut).
					WithTimeout(5 * time.Minute).
					WithPolling(15 * time.Second).
					Should(Equal(2))

				By(fmt.Sprintf("Scale in %s to 1 instance", appToScaleName))
				scaleIn := sendMetricToAutoscaler(cfg, appToScaleGUID, neighbourAppName, 100, true)
				Eventually(scaleIn).
					WithTimeout(5 * time.Minute).
					WithPolling(15 * time.Second).
					Should(Equal(1))

			})
		})
		Context("neighbour app send metrics if metrics strategy is not set i.e same_app", func() {
			BeforeEach(func() {
				policy = GenerateBindingsWithScalingPolicy("", 1, 2, "test_metric", 100, 550)
			})
			When("policy is attached with neighbour app", func() {
				It("should scale out and scale the neighbour app", func() {
					By(fmt.Sprintf("Scale out %s to 2 instance", neighbourAppName))
					scaleOut := sendMetricToAutoscaler(cfg, neighbourAppGUID, neighbourAppName, 550, true)
					Eventually(scaleOut).
						WithTimeout(5 * time.Minute).
						WithPolling(15 * time.Second).
						Should(Equal(2))

					By(fmt.Sprintf("Scale in %s to 1 instance", neighbourAppName))
					scaleIn := sendMetricToAutoscaler(cfg, neighbourAppGUID, neighbourAppName, 90, true)
					Eventually(scaleIn).
						WithTimeout(5 * time.Minute).
						WithPolling(15 * time.Second).
						Should(Equal(1))

				})
			})
			When("no policy is attached with neighbour app", func() {
				BeforeEach(func() {
					policy = ""
				})
				It("should not scale neighbour app", func() {
					sendMetricToAutoscaler(cfg, neighbourAppGUID, neighbourAppName, 550, true)
					Expect(RunningInstances(neighbourAppGUID, 5*time.Second)).To(Equal(1))

				})
			})

		})
	})
})

func sendMetricToAutoscaler(config *config.Config, appToScaleGUID string, neighbourAppName string, metricThreshold int, mtls bool) func() (int, error) {
	return func() (int, error) {

		if mtls {
			SendMetricMTLS(config, appToScaleGUID, neighbourAppName, metricThreshold)
		} else {
			SendMetric(config, neighbourAppName, metricThreshold)
		}
		return RunningInstances(appToScaleGUID, 5*time.Second)
	}
}
