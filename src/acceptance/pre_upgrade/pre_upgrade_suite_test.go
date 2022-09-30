package pre_upgrade_test

import (
	"acceptance/config"
	"acceptance/helpers"
	"testing"

	cth "github.com/KevinJCross/cf-test-helpers/v2/helpers"

	"github.com/KevinJCross/cf-test-helpers/v2/workflowhelpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	cfg   *config.Config
	setup *workflowhelpers.ReproducibleTestSuiteSetup
)

const componentName = "Pre Upgrade Test Suite"

func TestSetup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, componentName)
}

var _ = BeforeSuite(func() {
	cfg = config.LoadConfig()
	if cfg.GetArtifactsDirectory() != "" {
		cth.EnableCFTrace(cfg, componentName)
	}

	setup = workflowhelpers.NewTestSuiteSetup(cfg)

	GinkgoWriter.Println("Clearing down existing test orgs/spaces...")
	setup = workflowhelpers.NewTestSuiteSetup(cfg)

	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		orgs := helpers.GetTestOrgs(cfg)
		for _, org := range orgs {
			helpers.DeleteOrg(cfg, org)
		}
	})

	GinkgoWriter.Println("Clearing down existing test orgs/spaces... Complete")
	setup.Setup()

	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		if cfg.ShouldEnableServiceAccess() {
			helpers.EnableServiceAccess(cfg, setup.GetOrganizationName())
		}
	})

	if cfg.IsServiceOfferingEnabled() {
		helpers.CheckServiceExists(cfg, setup.TestSpace.SpaceName(), cfg.ServiceName)
	}
})
