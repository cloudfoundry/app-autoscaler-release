package broker_test

import (
	"testing"

	"acceptance/config"
	. "acceptance/helpers"

	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	cfg   *config.Config
	setup *workflowhelpers.ReproducibleTestSuiteSetup
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)

	cfg = config.LoadConfig(t)
	componentName := "Broker Suite"

	if cfg.GetArtifactsDirectory() != "" {
		helpers.EnableCFTrace(cfg, componentName)
	}
	if cfg.IsServiceOfferingEnabled() {
		RunSpecs(t, componentName)
	}
}

var _ = BeforeSuite(func() {

	setup = workflowhelpers.NewTestSuiteSetup(cfg)
	setup.Setup()

	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		if cfg.ShouldEnableServiceAccess() {
			EnableServiceAccess(cfg, setup.GetOrganizationName())
		}
	})

	CheckServiceExists(cfg)
})

var _ = AfterSuite(func() {
	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		if cfg.ShouldEnableServiceAccess() {
			DisableServiceAccess(cfg, setup.GetOrganizationName())
		}
	})
	setup.Teardown()
})
