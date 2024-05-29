package app_test

import (
	. "acceptance/helpers"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AutoScaler CF metadata support", func() {
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

	When("the label app-autoscaler.cloudfoundry.org/disable-autoscaling is set", func() {
		It("should not scale out", func() {
			By("Set the label app-autoscaler.cloudfoundry.org/disable-autoscaling to true")
			SetLabel(cfg, appGUID, "app-autoscaler.cloudfoundry.org/disable-autoscaling", "true")
			scaleOut := sendMetricToAutoscaler(cfg, appGUID, appName, 550, true)
			Consistently(scaleOut).
				WithTimeout(5 * time.Minute).
				WithPolling(15 * time.Second).
				Should(Equal(1))
		})
	})
})
