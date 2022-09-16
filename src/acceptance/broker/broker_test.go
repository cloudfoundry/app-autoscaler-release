package broker_test

import (
	"acceptance/helpers"
	"encoding/json"
	"fmt"
	url2 "net/url"
	"os"
	"strings"

	"github.com/KevinJCross/cf-test-helpers/v2/cf"
	"github.com/KevinJCross/cf-test-helpers/v2/generator"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

type serviceInstance string

func createService(onPlan string) serviceInstance {
	instanceName := generator.PrefixedRandomName(cfg.Prefix, cfg.InstancePrefix)
	By(fmt.Sprintf("create service %s on plan %s", instanceName, onPlan))
	createService := cf.Cf("create-service", cfg.ServiceName, onPlan, instanceName, "-b", cfg.ServiceBroker).Wait(cfg.DefaultTimeoutDuration())
	ExpectWithOffset(1, createService).To(Exit(0), "failed creating service")
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

func (s serviceInstance) delete() {
	deleteService := cf.Cf("delete-service", string(s), "-f").Wait(cfg.DefaultTimeoutDuration())
	Expect(deleteService).To(Exit(0))
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

	It("performs lifecycle operations", func() {
		instanceName := generator.PrefixedRandomName(cfg.Prefix, cfg.InstancePrefix)

		createService := cf.Cf("create-service", cfg.ServiceName, cfg.ServicePlan, instanceName, "-b", cfg.ServiceBroker).Wait(cfg.DefaultTimeoutDuration())
		Expect(createService).To(Exit(0), "failed creating service")

		bindService := cf.Cf("bind-service", appName, instanceName, "-c", "../assets/file/policy/invalid.json").Wait(cfg.DefaultTimeoutDuration())
		Expect(bindService).To(Exit(1))

		combinedBuffer := gbytes.BufferWithBytes(append(bindService.Out.Contents(), bindService.Err.Contents()...))
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

	It("should update service instance from autoscaler-free-plan to acceptance-standard", func() {
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

	It("should fail to update service instance from acceptance-standard to first", func() {
		plans := getPlans()
		if plans.length() < 2 {
			Skip(fmt.Sprintf("2 plans needed, only one plan available plans:%+v", plans))
			return
		}
		if !plans.contains("acceptance-standard") {
			Skip(fmt.Sprintf("Acceptance test standard plan required plans:%+v", plans))
			return
		}

		service := createService("acceptance-standard")
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
func (p plans) contains(planName string) bool {
	for _, plan := range p {
		if plan == planName {
			return true
		}
	}
	return false
}

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
