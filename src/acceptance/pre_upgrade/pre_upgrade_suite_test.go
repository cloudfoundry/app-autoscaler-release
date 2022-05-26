package pre_upgrade_test

import (
	"acceptance/config"
	"acceptance/helpers"
	"testing"

	"github.com/KevinJCross/cf-test-helpers/v2/workflowhelpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	cfg   *config.Config
	setup *workflowhelpers.ReproducibleTestSuiteSetup
)

func TestSetup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pre Upgrade Test Suite")
}

var _ = SynchronizedBeforeSuite(
	func() []byte {
		cfg = config.LoadConfig(GinkgoT())
		setup := workflowhelpers.NewTestSuiteSetup(cfg)
		helpers.Cleanup(cfg, setup)
		return nil
	},
	func([]byte) {
		cfg = config.LoadConfig(GinkgoT())
		setup = workflowhelpers.NewTestSuiteSetup(cfg)
		setup.Setup()

		workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
			if cfg.ShouldEnableServiceAccess() {
				helpers.EnableServiceAccess(cfg, setup.GetOrganizationName())
			}
		})

		if cfg.IsServiceOfferingEnabled() {
			helpers.CheckServiceExists(cfg, setup.TestSpace.SpaceName(), cfg.ServiceName)
		}
	},
)
