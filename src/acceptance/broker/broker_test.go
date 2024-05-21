package broker_test

import (
	"acceptance/helpers"
	"encoding/json"
	"fmt"
	url2 "net/url"
	"os"
	"strings"

	"github.com/cloudfoundry/cf-test-helpers/v2/generator"

	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry/cf-test-helpers/v2/cf"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type serviceInstance string

func createService(onPlan string) serviceInstance {
	instanceName := generator.PrefixedRandomName(cfg.Prefix, cfg.InstancePrefix)
	helpers.FailOnError(helpers.CreateServiceWithPlan(cfg, onPlan, instanceName))
	return serviceInstance(instanceName)
}

func createServiceWithParameters(onPlan string, parameters string) serviceInstance {
	instanceName := generator.PrefixedRandomName(cfg.Prefix, cfg.InstancePrefix)
	helpers.FailOnError(helpers.CreateServiceWithPlanAndParameters(cfg, onPlan, parameters, instanceName))
	return serviceInstance(instanceName)
}

func (s serviceInstance) updatePlan(toPlan string) {
	updateService := s.updatePlanRaw(toPlan)
	ExpectWithOffset(1, updateService).To(Exit(0), "failed updating service")
	Expect(strings.Contains(string(updateService.Out.Contents()), "The service does not support changing plans.")).To(BeFalse())
}

func (s serviceInstance) updatePlanRaw(toPlan string) *Session {
	By(fmt.Sprintf("update service plan to %s", toPlan))
	updateService := cf.Cf("update-service", string(s), "-p", toPlan).Wait(cfg.DefaultTimeoutDuration())
	return updateService
}

func (s serviceInstance) unbind(fromApp string) {
	unbindService := cf.Cf("unbind-service", fromApp, s.name()).Wait(cfg.DefaultTimeoutDuration())
	Expect(unbindService).To(Exit(0), "failed unbinding service instance %s from app %s", s.name(), fromApp)
}

func (s serviceInstance) delete() {
	deleteService := cf.Cf("delete-service", string(s), "-f").Wait(cfg.DefaultTimeoutDuration())
	Expect(deleteService).To(Exit(0), "failed deleting service instance %s", s.name())
}

func (s serviceInstance) name() string {
	return string(s)
}

var _ = Describe("AutoScaler Service Broker", func() {
	var appName string

	BeforeEach(func() {
		appName = helpers.CreateTestApp(cfg, "broker-test", 1)
	})

	AfterEach(func() {
		if os.Getenv("SKIP_TEARDOWN") == "true" {
			fmt.Println("Skipping Teardown...")
		} else {
			Eventually(cf.Cf("app", appName, "--guid"), cfg.DefaultTimeoutDuration()).Should(Exit())
			Eventually(cf.Cf("logs", appName, "--recent"), cfg.DefaultTimeoutDuration()).Should(Exit())
			Expect(cf.Cf("delete", appName, "-f", "-r").Wait(cfg.CfPushTimeoutDuration())).To(Exit(0))
		}
	})

	Context("performs lifecycle operations", func() {

		var instance serviceInstance

		BeforeEach(func() {
			instance = createService(cfg.ServicePlan)
		})

		It("fails to bind with invalid policies", func() {
			bindService := cf.Cf("bind-service", appName, instance.name(), "-c", "../assets/file/policy/invalid.json").Wait(cfg.DefaultTimeoutDuration())
			Expect(bindService).To(Exit(1))
			combinedBuffer := gbytes.BufferWithBytes(append(bindService.Out.Contents(), bindService.Err.Contents()...))
			Eventually(string(combinedBuffer.Contents())).Should(ContainSubstring(`[{"context":"(root).scaling_rules.1.adjustment","description":"Does not match pattern '^[-+][1-9]+[0-9]*%?$'"}]`))
		})

		It("binds&unbinds with policy", func() {
			policyFile := "../assets/file/policy/all.json"
			policy, err := os.ReadFile(policyFile)
			Expect(err).NotTo(HaveOccurred())

			err = helpers.BindServiceToAppWithPolicy(cfg, appName, instance.name(), policyFile)
			Expect(err).NotTo(HaveOccurred())

			bindingParameters := helpers.GetServiceCredentialBindingParameters(cfg, instance.name(), appName)
			Expect(bindingParameters).Should(MatchJSON(policy))

			instance.unbind(appName)
		})

		It("bind&unbinds without policy", func() {
			helpers.BindServiceToApp(cfg, appName, instance.name())
			bindingParameters := helpers.GetServiceCredentialBindingParameters(cfg, instance.name(), appName)
			Expect(bindingParameters).Should(MatchJSON("{}"))
			instance.unbind(appName)
		})

		AfterEach(func() {
			instance.delete()
		})
	})

	Describe("allows setting default policies", func() {
		var instance serviceInstance
		var defaultPolicy []byte
		var policy []byte

		BeforeEach(func() {
			instance = createServiceWithParameters(cfg.ServicePlan, "../assets/file/policy/default_policy.json")
			Expect(instance).NotTo(BeEmpty())
			var err error
			defaultPolicy, err = os.ReadFile("../assets/file/policy/default_policy.json")
			Expect(err).NotTo(HaveOccurred())

			var serviceParameters = struct {
				DefaultPolicy interface{} `json:"default_policy"`
			}{}

			err = json.Unmarshal(defaultPolicy, &serviceParameters)
			Expect(err).NotTo(HaveOccurred())

			policy, err = json.Marshal(serviceParameters.DefaultPolicy)
			Expect(err).NotTo(HaveOccurred())
		})

		It("allows retrieving the default policy using the Cloud Controller", func() {
			instanceParameters := helpers.GetServiceInstanceParameters(cfg, instance.name())
			Expect(instanceParameters).To(MatchJSON(defaultPolicy))
		})

		It("sets the default policy if no policy is set during binding and allows retrieving the policy via the binding parameters", func() {
			helpers.BindServiceToApp(cfg, appName, instance.name())

			bindingParameters := helpers.GetServiceCredentialBindingParameters(cfg, instance.name(), appName)
			Expect(bindingParameters).Should(MatchJSON(policy))

			unbindService := cf.Cf("unbind-service", appName, instance.name()).Wait(cfg.DefaultTimeoutDuration())
			Expect(unbindService).To(Exit(0), "failed unbinding service from app")

		})

		AfterEach(func() {
			if os.Getenv("SKIP_TEARDOWN") == "true" {
				fmt.Println("Skipping Teardown...")
			} else {
				instance.delete()
			}
		})
	})

	It("should update a service instance from the first plan to the second plan", func() {
		plans := getPlans()
		if plans.length() < 2 {
			Skip(fmt.Sprintf("2 plans needed, only one plan available plans:%+v", plans))
			return
		}
		service := createService(plans[0])
		service.updatePlan(plans[1])

		By("delete service")
		service.delete()
	})

	It("should fail to update a service instance from the second plan to the first plan", func() {
		plans := getPlans()
		if plans.length() < 2 {
			Skip(fmt.Sprintf("2 plans needed, only one plan available plans:%+v", plans))
			return
		}

		service := createService(plans[1])
		updateService := service.updatePlanRaw(plans[0])
		Expect(updateService).To(Exit(1), "failed updating service")

		errStream := updateService.Err
		if isCFVersion7() {
			errStream = updateService.Out
		}
		Expect(string(errStream.Contents())).To(ContainSubstring("service does not support changing plans"))
		service.delete()
	})
})

func isCFVersion7() bool {
	version := cf.Cf("--version").Wait(cfg.DefaultTimeoutDuration())
	Expect(version).To(Exit(0))
	return strings.Contains(string(version.Out.Contents()), "cf version 7")
}

type plans []string

func (p plans) length() int { return len(p) }

func getPlans() plans {
	values := url2.Values{
		"fields[service_offering.service_broker]": []string{"name"},
		"include":                []string{"service_offering"},
		"per_page":               []string{"5000"},
		"service_broker_names":   []string{cfg.ServiceBroker},
		"service_offering_names": []string{cfg.ServiceName},
	}
	url := &url2.URL{Path: "/v3/service_plans", RawQuery: values.Encode()}
	getPlans := cf.CfSilent("curl", url.String(), "-f").Wait(cfg.DefaultTimeoutDuration())
	Expect(getPlans).To(Exit(0), "failed getting plans")

	plansResult := &struct{ Resources []struct{ Name string } }{}
	err := json.Unmarshal(getPlans.Out.Contents(), plansResult)
	Expect(err).NotTo(HaveOccurred())

	var p plans
	for _, item := range plansResult.Resources {
		p = append(p, item.Name)
	}
	return p
}
