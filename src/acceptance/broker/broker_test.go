package broker_test

import (
	. "acceptance/broker"
	"acceptance/helpers"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("AutoScaler Service Broker", func() {
	var appName string

	BeforeEach(func() {
		appName = helpers.CreateTestApp(cfg, "broker-test", 1)
	})

	It("performs lifecycle operations", func() {
		broker := CreateInstance(cfg, OnPlan(cfg.ServicePlan))

		By("Try adding invalid policy json")
		bindResult := broker.BindAppRaw(appName, WithPolicy("../assets/file/policy/invalid.json")).Wait(cfg.DefaultTimeoutDuration())
		Expect(bindResult).To(Exit(1))

		combinedBuffer := gbytes.BufferWithBytes(append(bindResult.Out.Contents(), bindResult.Err.Contents()...))
		//Eventually(combinedBuffer).Should(gbytes.Say(`context":"(root).scaling_rules.1.adjustment","description":"Does not match pattern '^[-+][1-9]+[0-9]*$'"`))
		Eventually(string(combinedBuffer.Contents())).Should(ContainSubstring(`[{"context":"(root).scaling_rules.1.adjustment","description":"Does not match pattern '^[-+][1-9]+[0-9]*%?$'"}]`))

		By("Test bind&unbind with a policy")
		broker.BindApp(appName, WithPolicy("../assets/file/policy/all.json")).UnbindApp(appName)

		By("Test bind&unbind without policy")
		broker.BindApp(appName).UnbindApp(appName)

		broker.Delete()
	})

	It("should update service instance from  autoscaler-free-plan to acceptance-standard", func() {
		plans := GetPlans(cfg)
		if plans.Length() < 2 {
			Skip(fmt.Sprintf("2 plans needed, only one plan available plans:%+v", plans))
			return
		}
		service := CreateInstance(cfg, OnPlan(plans[0])).
			UpdatePlan(plans[1])

		By("delete service")
		service.Delete()
	})

	It("should fail to update service instance from acceptance-standard to first", func() {
		plans := GetPlans(cfg)
		if plans.Length() < 2 {
			Skip(fmt.Sprintf("2 plans needed, only one plan available plans:%+v", plans))
			return
		}
		if !plans.Contains("acceptance-standard") {
			Skip(fmt.Sprintf("Acceptance test standard plan required plans:%+v", plans))
			return
		}

		service := CreateInstance(cfg, OnPlan("acceptance-standard"))
		updateService := service.UpdatePlanRaw(plans[0])
		Expect(updateService).To(Exit(1), "failed updating service")
		Expect(strings.Contains(string(updateService.Out.Contents()), "The service does not support changing plans.")).To(BeTrue())

		service.Delete()
	})
})
