package broker

import (
	"acceptance/config"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type Instance struct {
	serviceName  string
	instanceName string
	brokerName   string
	plan         string
	timeout      time.Duration
}

type InstanceParam = func(*Instance) *Instance

func CreateInstance(cfg *config.Config, extraParams ...InstanceParam) *Instance {
	service := &Instance{
		instanceName: generator.PrefixedRandomName(cfg.Prefix, cfg.InstancePrefix),
		brokerName:   cfg.ServiceBroker,
		timeout:      cfg.DefaultTimeoutDuration()}
	for _, function := range extraParams {
		service = function(service)
	}
	return create(service)
}

func OnPlan(plan string) InstanceParam {
	return func(instance *Instance) *Instance {
		instance.plan = plan
		return instance
	}
}

func create(service *Instance) *Instance {
	By(fmt.Sprintf("create service %s on plan %s", service.instanceName, service.plan))
	createService := cf.Cf("create-service", service.serviceName, service.plan, service.instanceName, "-b", service.brokerName).Wait(service.timeout)
	Expect(createService).To(Exit(0), "failed creating service")
	return service
}

func (s *Instance) UpdatePlan(toPlan string) *Instance {
	updateService := s.UpdatePlanRaw(toPlan)
	Expect(updateService).To(Exit(0), "failed updating service")
	Expect(strings.Contains(string(updateService.Out.Contents()), "The service does not support changing plans.")).To(BeFalse())
	return s
}

func (s *Instance) UpdatePlanRaw(toPlan string) *Session {
	By(fmt.Sprintf("update service plan to %s", toPlan))
	updateService := cf.Cf("update-service", s.instanceName, "-p", toPlan).
		Wait(s.timeout)
	s.plan = toPlan
	return updateService
}

func (s Instance) Delete() {
	deleteService := cf.Cf("delete-service", s.instanceName, "-f").Wait(s.timeout)
	Expect(deleteService).To(Exit(0))
}

func (s Instance) Name() string {
	return s.instanceName
}

func (s Instance) UnbindApp(appName string) {
	Expect(cf.Cf("unbind-service", appName, s.instanceName).Wait(s.timeout)).
		To(Exit(0), "failed unbinding service from app")
}

type BindParam = func(Instance) []string

func WithPolicy(policyFile string) BindParam {
	return func(instance Instance) []string {
		return []string{"-c", policyFile}
	}
}

func (s Instance) BindApp(appName string, extraParams ...BindParam) Instance {
	Expect(s.BindAppRaw(appName, extraParams...).Wait(s.timeout)).
		To(Exit(0), "failed binding service to app without policy")
	return s
}

func (s Instance) BindAppRaw(appName string, extraParams ...BindParam) *Session {
	params := []string{"bind-service", appName, s.instanceName}
	for _, fn := range extraParams {
		params = append(params, fn(s)...)
	}
	return cf.Cf(params...)
}

type Plans []string

func (p Plans) Length() int { return len(p) }
func (p Plans) Contains(planName string) bool {
	for _, plan := range p {
		if plan == planName {
			return true
		}
	}
	return false
}

func GetPlans(cfg *config.Config) Plans {
	brokerName := "autoscaler"
	serviceOfferName := "autoscaler"
	getPlans := cf.Cf("curl",
		fmt.Sprintf("/v3/service_plans?fields[service_offering.service_broker]=name&service_broker_names=%s&service_offering_names=%s",
			brokerName, serviceOfferName),
		"-f").
		Wait(cfg.DefaultTimeoutDuration())
	Expect(getPlans).To(Exit(0), "failed getting plans")

	plansResult := &struct{ Resources []struct{ Name string } }{}
	err := json.Unmarshal(getPlans.Out.Contents(), plansResult)
	Expect(err).NotTo(HaveOccurred())

	var p Plans
	for _, item := range plansResult.Resources {
		p = append(p, item.Name)
	}
	return p
}
