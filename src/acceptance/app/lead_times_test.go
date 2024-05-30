package app_test

import (
	. "acceptance/helpers"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Autoscaler lead times for scaling", func() {
	var (
		policy string
		err    error
	)
	BeforeEach(func() {
		policy = GenerateDynamicScaleOutAndInPolicy(1, 2, "test_metric", 500, 500)
		appName = CreateTestApp(cfg, "labeled-go_app", 1)
		appGUID, err = GetAppGuid(cfg, appName)
		Expect(err).NotTo(HaveOccurred())
		instanceName = CreatePolicy(cfg, appName, appGUID, policy)
		StartApp(appName, cfg.CfPushTimeoutDuration())
	})
	AfterEach(AppAfterEach)

	When("lead times are defined", func() {
		FIt("should do first scaling after breach_duration_secs has passed and second scaling after cool_down_secs has passed", func() {
			breachDuration := TestBreachDurationSeconds * time.Second
			coolDown := TestCoolDownSeconds * time.Second
			internalMetricPollingIntervalOfAutoscaler := 40 * time.Second
			headroom := 60 * time.Second // be friendly and allow more time for "internal autoscaler processes" to happen before actual scaling is being done
			sendMetricForScaleOutAndReturnNumInstancesFunc := sendMetricToAutoscaler(cfg, appGUID, appName, 510, false)
			sendMetricForScaleInAndReturnNumInstancesFunc := sendMetricToAutoscaler(cfg, appGUID, appName, 490, false)

			By("checking that no scaling out happens before breach_duration_secs have passed")
			Consistently(sendMetricForScaleOutAndReturnNumInstancesFunc).
				WithTimeout(breachDuration).
				WithPolling(time.Second).
				Should(Equal(1))

			By("checking that scale out happens after breach_duration_secs have passed")
			Eventually(sendMetricForScaleOutAndReturnNumInstancesFunc).
				WithTimeout(internalMetricPollingIntervalOfAutoscaler + headroom).
				WithPolling(time.Second).
				Should(Equal(2))

			By("checking that no scale in happens before breach_duration_secs and cool_down_secs have passed")
			Consistently(sendMetricForScaleInAndReturnNumInstancesFunc).
				WithTimeout(breachDuration + coolDown).
				WithPolling(time.Second).
				Should(Equal(2))

			By("checking that scale in happens after breach_duration_secs and cool_down_secs have passed")
			Eventually(sendMetricForScaleInAndReturnNumInstancesFunc).
				WithTimeout(internalMetricPollingIntervalOfAutoscaler + headroom).
				WithPolling(time.Second).
				Should(Equal(1))
		})

	})
})
