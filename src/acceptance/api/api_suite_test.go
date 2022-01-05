package api_test

import (
	"acceptance/config"
	. "acceptance/helpers"
	"fmt"
	"os"
	"testing"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var (
	app          *App
	cfg          *config.Config
	instanceName string
	setup        *workflowhelpers.ReproducibleTestSuiteSetup
	otherSetup   *workflowhelpers.ReproducibleTestSuiteSetup
	healthURL    string
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)

	cfg = config.LoadConfig(t)
	componentName := "Public API Suite"

	if cfg.GetArtifactsDirectory() != "" {
		helpers.EnableCFTrace(cfg, componentName)
	}

	RunSpecs(t, componentName)
}

var _ = BeforeSuite(func() {

	otherConfig := *cfg
	otherConfig.NamePrefix = otherConfig.NamePrefix + "_other"

	setup = workflowhelpers.NewTestSuiteSetup(cfg)
	otherSetup = workflowhelpers.NewTestSuiteSetup(&otherConfig)

	Cleanup(cfg, setup)
	Cleanup(&otherConfig, otherSetup)

	otherSetup.Setup()
	setup.Setup()

	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		if cfg.IsServiceOfferingEnabled() && cfg.ShouldEnableServiceAccess() {
			EnableServiceAccess(cfg, setup.GetOrganizationName())
		}
	})

	app = New(cfg)
	app.Create(1)
	app.Start()
	healthURL = fmt.Sprintf("%s%s", cfg.ASApiEndpoint, HealthPath)
	if cfg.IsServiceOfferingEnabled() {
		CheckServiceExists(cfg)

		instanceName = generator.PrefixedRandomName(cfg.Prefix, cfg.InstancePrefix)
		createService := cf.Cf("create-service", cfg.ServiceName, cfg.ServicePlan, instanceName, "-b", cfg.ServiceBroker).Wait(cfg.DefaultTimeoutDuration())
		Expect(createService).To(Exit(0), fmt.Sprintf("failed creating service %s", instanceName))

		bindService := cf.Cf("bind-service", app.name, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(bindService).To(Exit(0), fmt.Sprintf("failed binding service %s to app %s", instanceName, app.name))
	}
})

var _ = AfterSuite(func() {
	if os.Getenv("SKIP_TEARDOWN") == "true" {
		fmt.Println("Skipping Teardown...")
	} else {
		if cfg.IsServiceOfferingEnabled() {
			if app != nil && instanceName != "" {
				unbindService := cf.Cf("unbind-service", app.name, instanceName).Wait(cfg.DefaultTimeoutDuration())
				if unbindService.ExitCode() != 0 {
					purgeService := cf.Cf("purge-service-instance", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
					Expect(purgeService).To(Exit(0), fmt.Sprintf("failed to purge service instance %s", instanceName))
				}
			}

			if instanceName != "" {
				deleteService := cf.Cf("delete-service", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
				if deleteService.ExitCode() != 0 {
					purgeService := cf.Cf("purge-service-instance", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
					Expect(purgeService).To(Exit(0), fmt.Sprintf("failed to purge service instance %s", instanceName))
				}
			}
		}

		app.Delete()

		workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
			if cfg.IsServiceOfferingEnabled() && cfg.ShouldEnableServiceAccess() {
				DisableServiceAccess(cfg, setup.GetOrganizationName())
			}
		})

		otherSetup.Teardown()
		setup.Teardown()
	}
})
