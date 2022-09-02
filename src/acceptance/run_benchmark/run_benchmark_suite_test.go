package run_benchmark

import (
	"acceptance/config"
	"acceptance/helpers"
	"fmt"
	"os"
	"testing"

	"github.com/KevinJCross/cf-test-helpers/v2/workflowhelpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	cfg   *config.Config
	setup *workflowhelpers.ReproducibleTestSuiteSetup
	orgName   string
	orgGUID   string
	spaceName string
	spaceGUID string
)

func TestSetup(t *testing.T) {
	RegisterFailHandler(Fail)
	cfg = config.LoadConfig(t)
	setup = workflowhelpers.NewTestSuiteSetup(cfg)
	RunSpecs(t, "Pre Upgrade Test Suite")
}

var _ = BeforeSuite(func() {
	// use smoke test to avoid creating a new user
	setup = workflowhelpers.NewSmokeTestSuiteSetup(cfg)

	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		orgs := helpers.GetTestOrgs(cfg)
		Expect(len(orgs)).To(Equal(1))
		orgName = orgs[0]
		_, orgGUID, spaceName, spaceGUID = helpers.GetOrgSpaceNamesAndGuids(cfg, orgName)
	})

	Expect(orgName).ToNot(Equal(""), "orgName has not been determined")
	Expect(spaceName).ToNot(Equal(""), "spaceName has not been determined")

	// discover the org / space from the environment
	cfg.UseExistingOrganization = true
	cfg.UseExistingSpace = true

	cfg.ExistingOrganization = orgName
	cfg.ExistingSpace = spaceName

	setup = workflowhelpers.NewTestSuiteSetup(cfg)

	setup.Setup()

	if cfg.IsServiceOfferingEnabled() {
		helpers.CheckServiceExists(cfg, setup.TestSpace.SpaceName(), cfg.ServiceName)
	}
})

var _ = AfterSuite(func() {
	if os.Getenv("SKIP_TEARDOWN") == "true" {
		fmt.Println("Skipping Teardown...")
	} else {
		// TODO: Cleanup test
	}
})
