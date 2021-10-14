package pre_upgrade_test

import (
	. "acceptance/helpers"

	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AutoScaler dynamic policy", func() {
	var (
		appName              string
		appGUID              string
		policy               string
		initialInstanceCount = 1
	)

	JustBeforeEach(func() {
		appName = CreateTestApp(cfg, "nodeapp-cpu", initialInstanceCount)
		appGUID = GetAppGuid(cfg, appName)
		StartApp(appName, cfg.CfPushTimeoutDuration())
		WaitForNInstancesRunning(appGUID, initialInstanceCount, cfg.DefaultTimeoutDuration())
		_ = CreatePolicy(cfg, appName, appGUID, policy)
	})

	Context("when scaling by cpu", func() {

		Context("when cpu is greater than and then less than threshold", func() {

			BeforeEach(func() {
				policy = GenerateDynamicScaleOutAndInPolicy(1, 2, "cpu", 5, 10)
				initialInstanceCount = 1
			})

			It("should scale out and back in", func() {
				response := helpers.CurlAppWithTimeout(cfg, appName, "/cpu/50/1", 10*time.Second)
				Expect(response).Should(ContainSubstring(`set app cpu utilization to 50% for 1 minutes, busyTime=10, idleTime=10`))

				WaitForNInstancesRunning(appGUID, 2, 3*time.Minute)

				// lets attempt to scale back down
				response = helpers.CurlAppWithTimeout(cfg, appName, "/cpu/0/5", 10*time.Second)
				Expect(response).Should(ContainSubstring(`set app cpu utilization to 1% for 5 minutes, busyTime=10, idleTime=990`))

				WaitForNInstancesRunning(appGUID, 1, 3*time.Minute)
			})

		})

	})

})
