package post_upgrade_test

import (
	. "acceptance/helpers"
	"strings"

	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

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
		apps := GetApps(cfg, orgGUID, spaceGUID, "autoscaler-")
		for _, app := range apps {
			if strings.Contains(app, "nodeapp-cpu") {
				appName = app
				appGUID = GetAppGuid(cfg, appName)
			}
		}
		Expect(appName).ShouldNot(Equal(""), "Unable to determine nodeapp-cpu from space")
	})

	Context("when scaling by cpu", func() {

		Context("when cpu is greater than scaling out threshold", func() {

			It("should have a policy attached", func() {
				policy := GetPolicy(cfg, appGUID)
				Expect(policy.InstanceMin).To(Equal(1))
				Expect(policy.InstanceMax).To(Equal(2))
				Expect(len(policy.ScalingRules)).To(Equal(2))

				Expect(policy.ScalingRules[0].MetricType).To(Equal("cpu"))
				Expect(policy.ScalingRules[0].Threshold).To(Equal(int64(10)))
				Expect(policy.ScalingRules[0].Operator).To(Equal(">="))
				Expect(policy.ScalingRules[0].CoolDownSeconds).To(Equal(TestCoolDownSeconds))
				Expect(policy.ScalingRules[0].BreachDurationSeconds).To(Equal(TestBreachDurationSeconds))
				Expect(policy.ScalingRules[0].Adjustment).To(Equal("+1"))

				Expect(policy.ScalingRules[1].MetricType).To(Equal("cpu"))
				Expect(policy.ScalingRules[1].Threshold).To(Equal(int64(2)))
				Expect(policy.ScalingRules[1].Operator).To(Equal("<"))
				Expect(policy.ScalingRules[1].CoolDownSeconds).To(Equal(TestCoolDownSeconds))
				Expect(policy.ScalingRules[1].BreachDurationSeconds).To(Equal(TestBreachDurationSeconds))
				Expect(policy.ScalingRules[1].Adjustment).To(Equal("-1"))
			})

			It("should scale out and in again", func() {
				totalTime := time.Duration(interval*2)*time.Second + 1*time.Minute
				finishTime := time.Now().Add(totalTime)

				WaitForNInstancesRunning(appGUID, 1, time.Until(finishTime))

				response := helpers.CurlAppWithTimeout(cfg, appName, "/cpu/50/1", 10*time.Second)
				Expect(response).Should(ContainSubstring(`set app cpu utilization to 50% for 1 minutes, busyTime=10, idleTime=10`))

				totalTime = time.Duration(interval*2)*time.Second + 3*time.Minute
				finishTime = time.Now().Add(totalTime)

				Eventually(func() float64 {
					return AverageCPUByInstance(appGUID, totalTime)
				}, totalTime, 15*time.Second).Should(BeNumerically(">=", 0.02))

				WaitForNInstancesRunning(appGUID, 2, time.Until(finishTime))

				// lets attempt to scale back down
				response = helpers.CurlAppWithTimeout(cfg, appName, "/cpu/0/5", 10*time.Second)
				Expect(response).Should(ContainSubstring(`set app cpu utilization to 1% for 5 minutes, busyTime=10, idleTime=990`))

				totalTime = time.Duration(interval*2)*time.Second + 10*time.Minute
				finishTime = time.Now().Add(totalTime)

				Eventually(func() float64 {
					return AverageCPUByInstance(appGUID, totalTime)
				}, totalTime, 15*time.Second).Should(BeNumerically("<=", 0.1))

				WaitForNInstancesRunning(appGUID, 1, time.Until(finishTime))
			})

		})

	})

})
