package broker_test

import (
	"acceptance/helpers"
	"strings"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
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

	AfterEach(func() {
		Eventually(cf.Cf("app", appName, "--guid"), cfg.DefaultTimeoutDuration()).Should(Exit())
		Eventually(cf.Cf("logs", appName, "--recent"), cfg.DefaultTimeoutDuration()).Should(Exit())
		Expect(cf.Cf("delete", appName, "-f", "-r").Wait(cfg.CfPushTimeoutDuration())).To(Exit(0))
	})

	It("performs lifecycle operations", func() {
		instanceName := generator.PrefixedRandomName("autoscaler", "service")

		createService := cf.Cf("create-service", cfg.ServiceName, cfg.ServicePlan, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(createService).To(Exit(0), "failed creating service")

		bindService := cf.Cf("bind-service", appName, instanceName, "-c", "../assets/file/policy/invalid.json").Wait(cfg.DefaultTimeoutDuration())
		Expect(bindService).To(Exit(1))

		combinedBuffer := gbytes.BufferWithBytes(append(bindService.Out.Contents(), bindService.Err.Contents()...))
		//Eventually(combinedBuffer).Should(gbytes.Say(`context":"(root).scaling_rules.1.adjustment","description":"Does not match pattern '^[-+][1-9]+[0-9]*$'"`))
		Eventually(string(combinedBuffer.Contents())).Should(ContainSubstring(`[{"context":"(root).scaling_rules.1.adjustment","description":"Does not match pattern '^[-+][1-9]+[0-9]*%?$'"}]`))

		By("Test bind&unbind with policy")
		bindService = cf.Cf("bind-service", appName, instanceName, "-c", "../assets/file/policy/all.json").Wait(cfg.DefaultTimeoutDuration())
		Expect(bindService).To(Exit(0), "failed binding service to app with a policy ")

		unbindService := cf.Cf("unbind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(unbindService).To(Exit(0), "failed unbinding service from app")

		By("Test bind&unbind without policy")
		bindService = cf.Cf("bind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(bindService).To(Exit(0), "failed binding service to app without policy")

		unbindService = cf.Cf("unbind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(unbindService).To(Exit(0), "failed unbinding service from app")

		deleteService := cf.Cf("delete-service", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
		Expect(deleteService).To(Exit(0))
	})

	FIt("should update service instance from  autoscaler-free-plan to acceptance-standard", func() {
		//TODO only run if there is  "autoscaler-free-plan" as first and "acceptance-standard" as second plan"
		instanceName := generator.PrefixedRandomName("autoscaler", "service")
		//TODO get first plan name
		servicePlanName := "autoscaler-free-plan"
		createService := cf.Cf("create-service", cfg.ServiceName, servicePlanName, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(createService).To(Exit(0), "failed creating service")

		//TODO get second plan name
		updateToServicePlanName := "acceptance-standard"
		updateService := cf.Cf("update-service", instanceName, "-p", updateToServicePlanName).Wait(cfg.DefaultTimeoutDuration())
		Expect(updateService).To(Exit(0), "failed updating service")

		Expect(strings.Contains(string(updateService.Out.Contents()), "The service does not support changing plans.")).To(BeFalse())

		deleteService := cf.Cf("delete-service", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
		Expect(deleteService).To(Exit(0))
	})

	It("should not update service instance from acceptance-standard to first", func() {
		//TODO only run if there is  "autoscaler-free-plan" as first and "acceptance-standard" as second plan"
		instanceName := generator.PrefixedRandomName("autoscaler", "service")
		servicePlanName := "acceptance-standard"
		createService := cf.Cf("create-service", cfg.ServiceName, servicePlanName, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(createService).To(Exit(0), "failed creating service")

		updateToServicePlanName := "autoscaler-free-plan"
		updateService := cf.Cf("update-service", instanceName, "-p", updateToServicePlanName).Wait(cfg.DefaultTimeoutDuration())
		Expect(updateService).To(Exit(1), "failed updating service")
		Expect(strings.Contains(string(updateService.Out.Contents()), "The service does not support changing plans.")).To(BeTrue())

		deleteService := cf.Cf("delete-service", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
		Expect(deleteService).To(Exit(0))
	})
})
