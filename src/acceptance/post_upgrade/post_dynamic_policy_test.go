package post_upgrade_test

import (
	"acceptance/helpers"
	"fmt"

	cfh "github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"
)

var _ = Describe("AutoScaler dynamic policy", func() {

	var (
		appName string
		appGUID string
	)

	JustBeforeEach(func() {
		appName, appGUID = GetAppInfo(orgGUID, spaceGUID, "nodeapp-cpu")
		Expect(appName).ShouldNot(Equal(""), "Unable to determine nodeapp-cpu from space")
	})

	// To check existing aggregated cpu metrics do: cf asm APP_NAME cpu
	// See: https://www.ibm.com/docs/de/cloud-private/3.2.0?topic=SSBS6K_3.2.0/cloud_foundry/integrating/cfee_autoscaler.html
	Context("when scaling by cpu", func() {
		It("when cpu is greater than scaling out threshold", func() {
			By("should have a policy attached")
			policy := helpers.GetPolicy(cfg, appGUID)
			expectedPolicy := helpers.ScalingPolicy{InstanceMin: 1, InstanceMax: 2,
				ScalingRules: []*helpers.ScalingRule{
					{MetricType: "cpu", BreachDurationSeconds: helpers.TestBreachDurationSeconds,
						Threshold: 10, Operator: ">=", Adjustment: "+1", CoolDownSeconds: helpers.TestCoolDownSeconds},
					{MetricType: "cpu", BreachDurationSeconds: helpers.TestBreachDurationSeconds,
						Threshold: 5, Operator: "<", Adjustment: "-1", CoolDownSeconds: helpers.TestCoolDownSeconds},
				},
			}
			Expect(expectedPolicy).To(BeEquivalentTo(policy))

			By("should scale out and in again")
			Expect(helpers.RunningInstances(appGUID, 5*time.Second)).To(Equal(1))
			helpers.WaitForNInstancesRunning(appGUID, 1, 3*time.Minute)

			response := cfh.CurlAppWithTimeout(cfg, appName, "/cpu/50/1", 10*time.Second)
			Expect(response).Should(ContainSubstring(`set app cpu utilization to 50% for 1 minutes, busyTime=10, idleTime=10`))

			helpers.WaitForNInstancesRunning(appGUID, 2, 3*time.Minute)

			By("lets attempt to scale back down")

			for i := 0; i < 2; i++ {
				response = cfh.CurlAppWithTimeout(cfg, appName, "/cpu/close", 10*time.Second, "-H", fmt.Sprintf(`X-Cf-App-Instance: %s:%d`, appGUID, i))
				Expect(response).Should(ContainSubstring(`close cpu test`))
			}

			helpers.WaitForNInstancesRunning(appGUID, 1, 3*time.Minute)
		})
	})
})
