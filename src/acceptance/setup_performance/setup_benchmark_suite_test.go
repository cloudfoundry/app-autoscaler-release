package pre_upgrade_test

import (
	"acceptance/config"
	"acceptance/helpers"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/KevinJCross/cf-test-helpers/v2/cf"
	"github.com/onsi/gomega/gexec"

	"github.com/KevinJCross/cf-test-helpers/v2/workflowhelpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	cfg                *config.Config
	setup              *workflowhelpers.ReproducibleTestSuiteSetup
	nodeAppDropletPath string
)

func TestSetup(t *testing.T) {
	RegisterFailHandler(Fail)
	cfg = config.LoadConfig()
	setup = workflowhelpers.NewTestSuiteSetup(cfg)

	RunSpecs(t, "Pre Upgrade Test Suite")
}

var _ = BeforeSuite(func() {

	fmt.Println("Clearing down existing test orgs/spaces...")
	// Infinite memory quota OO
	setup = workflowhelpers.NewRunawayAppTestSuiteSetup(cfg)

	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		orgs := helpers.GetTestOrgs(cfg)

		for _, org := range orgs {
			orgName, _, spaceName, _ := helpers.GetOrgSpaceNamesAndGuids(cfg, org)
			if spaceName != "" {
				helpers.DeleteOrgWithTimeout(cfg, orgName, time.Duration(120)*time.Second)
			}
		}
	})

	fmt.Println("Clearing down existing test orgs/spaces... Complete")
	setup.Setup()

	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		if cfg.ShouldEnableServiceAccess() {
			helpers.EnableServiceAccess(cfg, setup.GetOrganizationName())
		}
		orgGuid := helpers.GetOrgGuid(cfg, setup.GetOrganizationName())
		orgQuotaName := helpers.GetOrgQuotaNameFrom(orgGuid, cfg.DefaultTimeoutDuration())
		updateOrgQuota := cf.Cf("update-org-quota", orgQuotaName, "-m", strconv.Itoa(cfg.Performance.AppCount*256)+"MB", "-r", strconv.Itoa(cfg.Performance.AppCount*2), "-s", strconv.Itoa(cfg.Performance.AppCount*2)).Wait(cfg.DefaultTimeoutDuration())
		Expect(updateOrgQuota).To(gexec.Exit(0), "unable update org quota: "+string(updateOrgQuota.Out.Contents()[:]))
	})

	if cfg.IsServiceOfferingEnabled() {
		helpers.CheckServiceExists(cfg, setup.TestSpace.SpaceName(), cfg.ServiceName)
	}

	fmt.Println("creating droplet")
	nodeAppDropletPath = helpers.CreateDroplet(*cfg)

})
