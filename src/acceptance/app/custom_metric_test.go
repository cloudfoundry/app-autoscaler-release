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

var _ = Describe("AutoScaler custom metrics", func() {
	var (
		policy string
		err    error
	)
	BeforeEach(func() {

		appToScaleName = CreateTestApp(cfg, "go-custom-metric", 1)
		appToScaleGUID, err = GetAppGuid(cfg, appToScaleName)
		Expect(err).NotTo(HaveOccurred())

	})
	AfterEach(AppAfterEach)

	Describe("custom metrics policy for same app", func() {
		BeforeEach(func() {
			policy = GenerateDynamicScaleOutAndInPolicy(1, 2, "test_metric", 500, 500)
			instanceName = CreatePolicy(cfg, appToScaleName, appToScaleGUID, policy)
			StartApp(appToScaleName, cfg.CfPushTimeoutDuration())
		})
		// This test will fail if credential-type is set to X509 in autoscaler broker.
		// Therefore, only mtls connection will be supported for custom metrics in future
		Context("when scaling by custom metrics", func() {
			BeforeEach(func() {
				//instanceName = CreatePolicy(cfg, appToScaleName, appToScaleGUID, policy)
				//StartApp(appToScaleName, cfg.CfPushTimeoutDuration())
			})
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
			BeforeEach(func() {
				//instanceName = CreatePolicy(cfg, appToScaleName, appToScaleGUID, policy)
				//StartApp(appToScaleName, cfg.CfPushTimeoutDuration())
			})
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

	})

	Describe("Custom metrics with neighbour app", func() {
		BeforeEach(func() {
			// attach policy to appToScale B
			policy = GenerateBindingsWithScalingPolicy("bound_app", 1, 2, "test_metric", 100, 500)
			instanceName = CreatePolicy(cfg, appToScaleName, appToScaleGUID, policy)
			StartApp(appToScaleName, cfg.CfPushTimeoutDuration())

			// push neighbour app
			neighbourAppName = CreateTestApp(cfg, "go-custom_metric_producer-app", 1)
			neighbourAppGUID, err = GetAppGuid(cfg, neighbourAppName)
			Expect(err).NotTo(HaveOccurred())
			err := BindServiceToAppWithPolicy(cfg, neighbourAppName, instanceName, "")
			Expect(err).NotTo(HaveOccurred())
			StartApp(neighbourAppName, cfg.CfPushTimeoutDuration())

		})
		Context("neighbour app sends custom metrics for app B via mtls", func() {
			JustBeforeEach(func() {

			})

			FWhen("policy is attached with the appToScale with a bound_app mentioned", func() {
				BeforeEach(func() {
					policy = GenerateBindingsWithScalingPolicy("bound_app", 1, 2, "test_metric", 100, 500)
				})
				It("should scale out and scale in app B", Label(acceptance.LabelSmokeTests), func() {
					By(fmt.Sprintf("Scale out %s to 2 instance", appToScaleName))
					scaleOut := sendMetricToAutoscaler(cfg, appToScaleGUID, neighbourAppName, 550, true)
					Eventually(scaleOut).
						WithTimeout(5 * time.Minute).
						WithPolling(15 * time.Second).
						Should(Equal(2))

					By(fmt.Sprintf("Scale in %s to 1 instance", appToScaleName))
					scaleIn := sendMetricToAutoscaler(cfg, appToScaleGUID, neighbourAppName, 80, true)
					Eventually(scaleIn).
						WithTimeout(5 * time.Minute).
						WithPolling(15 * time.Second).
						Should(Equal(1))

				})
			})
			//FixME  ? Is the following valid?
			/*
					cf bind-service autoscaler-3-go-neighbour-app-25a4dc3fb9e6ea00
					autoscaler-3-service-64a8ea1ff7d7f3f6 -c
					{"configuration":{"custom_metrics":{"auth":{"credential_type":""},
					"metric_submission_strategy":{"allow_from":"bound_app"}}},
					"instance_min_count":0,"instance_max_count":0}

				Command
				 cf bind-service autoscaler-3-go-neighbour-app-25a4dc3fb9e6ea00 autoscaler-3-service-64a8ea1ff7d7f3f6 -c {"configuration":{"custom_metrics":{"auth":{"credential_type":""},"metric_submission_strategy":{"allow_from":"bound_app"}}},"instance_min_count":0,"instance_max_count":0}

				  Binding service instance autoscaler-3-service-64a8ea1ff7d7f3f6 to app autoscaler-3-go-neighbour-app-25a4dc3fb9e6ea00 in org arsalan / space test-space as autoscaler-3211-TESTS-3-USER-c41b39e65d2ce188...
				  Job (e3fee92d-1062-4853-a6d4-d017f6b43157) failed: bind could not be completed: Service broker error: invalid policy provided: [{"context":"(root)","description":"Must validate at least one schema (anyOf)"},{"context":"(root)","description":"scaling_rules is required"},{"context":"(root).instance_min_count","description":"Must be greater than or equal to 1"}]

			*/
			XWhen("policy is not attached with the neighbour app", func() {
				BeforeEach(func() {
					policy = GenerateBindingConfiguration("bound_app")
				})
				It("should scale out and scale in app B", func() {
					By(fmt.Sprintf("Scale out %s to 2 instance", appToScaleName))
					scaleOut := sendMetricToAutoscaler(cfg, appToScaleGUID, neighbourAppName, 550, true)
					Eventually(scaleOut).
						WithTimeout(5 * time.Minute).
						WithPolling(15 * time.Second).
						Should(Equal(2))

					By(fmt.Sprintf("Scale in %s to 1 instance", appToScaleName))
					scaleIn := sendMetricToAutoscaler(cfg, appToScaleGUID, neighbourAppName, 80, true)
					Eventually(scaleIn).
						WithTimeout(5 * time.Minute).
						WithPolling(15 * time.Second).
						Should(Equal(1))

				})
			})

			XWhen("app B tries to send metrics for neighbour app with strategy same_app", func() {
				BeforeEach(func() {
					policy = GenerateBindingsWithScalingPolicy("bound_app", 1, 2, "test_metric", 100, 500)
				})
				It("should not scale neighbour app", func() {
					By(fmt.Sprintf("Fail Scale %s ", neighbourAppName))
					sendMetricToAutoscaler(cfg, neighbourAppGUID, appToScaleName, 550, true)
					WaitForNInstancesRunning(neighbourAppGUID, 1, 5*time.Second, "expected 1 instance running")
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
