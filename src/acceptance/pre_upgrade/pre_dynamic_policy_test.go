package pre_upgrade_test

import (
	"acceptance/helpers"

	"time"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("AutoScaler dynamic policy", func() {
	var (
		appName              string
		appGUID              string
		policy               string
		initialInstanceCount = 1
	)

	JustBeforeEach(func() {
		appName = helpers.CreateTestApp(cfg, "nodeapp-cpu", initialInstanceCount)
		appGUID = helpers.GetAppGuid(cfg, appName)
		helpers.StartApp(appName, cfg.CfPushTimeoutDuration())
		helpers.WaitForNInstancesRunning(appGUID, initialInstanceCount, cfg.DefaultTimeoutDuration())
		_ = helpers.CreatePolicy(cfg, appName, appGUID, policy)
	})

	Context("when scaling by cpu", func() {

		Context("when cpu is greater than and then less than threshold", func() {

			BeforeEach(func() {
				policy = helpers.GenerateDynamicScaleOutAndInPolicy(1, 2, "cpu", 5, 10)
				initialInstanceCount = 1
			})

			It("should scale out and back in", func() {

				By("should scale out to 2 instances")
				helpers.AppSetCpuUsage(cfg, appName, 50, 5)
				helpers.WaitForNInstancesRunning(appGUID, 2, 10*time.Minute)

				By("should scale in to 1 instance after cpu usage is reduced")
				//only hit the one instance that was asked to run hot.
				helpers.AppEndCpuTest(cfg, appName, 0)
				helpers.WaitForNInstancesRunning(appGUID, 1, 10*time.Minute)
			})

		})

	})

})
