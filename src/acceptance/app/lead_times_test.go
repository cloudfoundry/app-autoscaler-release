package app_test

import (
	. "acceptance/helpers"
	"fmt"
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
			internalMetricPollingIntervalOfAutoscaler := 40 * time.Second
			headroom := 60 * time.Second // be patient and allow more time for "internal autoscaler processes" to happen before actual scaling happens

			breachDuration := TestBreachDurationSeconds * time.Second
			expectedFirstScalingTimeWindow := internalMetricPollingIntervalOfAutoscaler + breachDuration + headroom
			scaleOut := sendMetricToAutoscaler(cfg, appGUID, appName, 501, false)
			Eventually(scaleOut).
				WithTimeout(expectedFirstScalingTimeWindow).
				WithPolling(time.Second).
				Should(Equal(2))

			fmt.Println(time.Now())
			coolDown := TestCoolDownSeconds * time.Second
			expectedSecondScalingTimeWindow := internalMetricPollingIntervalOfAutoscaler + breachDuration + coolDown + headroom
			scaleIn := sendMetricToAutoscaler(cfg, appGUID, appName, 0, false)
			Eventually(scaleIn).
				WithTimeout(expectedSecondScalingTimeWindow).
				WithPolling(time.Second).
				Should(Equal(1))
			fmt.Println(time.Now())
		})
	})
})
