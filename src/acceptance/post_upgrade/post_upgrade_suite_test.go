package post_upgrade_test

import (
	"acceptance/config"
	"acceptance/helpers"
	"fmt"
	"os"
	"testing"

	"github.com/KevinJCross/cf-test-helpers/v2/cf"
	"github.com/KevinJCross/cf-test-helpers/v2/workflowhelpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var (
	cfg       *config.Config
	setup     *workflowhelpers.ReproducibleTestSuiteSetup
	orgName   string
	orgGUID   string
	spaceName string
	spaceGUID string
)

func TestSetup(t *testing.T) {
	RegisterFailHandler(Fail)
	cfg = config.LoadConfig(t)
	RunSpecs(t, "Post Upgrade Test Suite")
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
		helpers.CheckServiceExists(cfg)
	}

})

var _ = AfterSuite(func() {
	if os.Getenv("SKIP_TEARDOWN") == "true" {
		fmt.Println("Skipping Teardown...")
	} else {
		fmt.Println("Clearing down existing test orgs/spaces...")

		workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
			orgs := helpers.GetTestOrgs(cfg)

			for _, org := range orgs {
				orgName, orgGuid, spaceName, spaceGuid := helpers.GetOrgSpaceNamesAndGuids(cfg, org)
				if spaceName != "" {
					target := cf.Cf("target", "-o", orgName, "-s", spaceName).Wait(cfg.DefaultTimeoutDuration())
					Expect(target).To(Exit(0), fmt.Sprintf("failed to target %s and %s", orgName, spaceName))

					apps := helpers.GetApps(cfg, orgGuid, spaceGuid, "autoscaler-")
					helpers.DeleteApps(cfg, apps, 0)

					services := helpers.GetServices(cfg, orgGuid, spaceGuid, "autoscaler-")
					helpers.DeleteServices(cfg, services)
				}

				helpers.DeleteOrg(cfg, org)
			}
		})

		fmt.Println("Clearing down existing test orgs/spaces... Complete")
	}
})
