package peformance_setup_test

import (
	"acceptance/config"
	. "acceptance/helpers"
	"fmt"
	"github.com/KevinJCross/cf-test-helpers/v2/cf"
	"github.com/KevinJCross/cf-test-helpers/v2/workflowhelpers"
	"github.com/onsi/gomega/gexec"
	"strconv"
	"testing"
	"time"

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

	if cfg.Performance.Teardown {
		cleanup()

	}
	setup.Setup()

	EnableServiceAccess(setup, cfg, setup.GetOrganizationName())
	if cfg.Performance.UpdateExistingOrgQuota {
		workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
			orgGuid := GetOrgGuid(cfg, setup.GetOrganizationName())
			orgQuotaName := GetOrgQuotaNameFrom(orgGuid, cfg.DefaultTimeoutDuration())
			updateOrgQuota(orgQuotaName, cfg.Performance.AppCount, cfg.DefaultTimeoutDuration())
		})
	}

	if cfg.IsServiceOfferingEnabled() {
		CheckServiceExists(cfg, setup.TestSpace.SpaceName(), cfg.ServiceName)
	}

	fmt.Println("creating droplet")
	nodeAppDropletPath = CreateDroplet(*cfg)

})

func cleanup() {
	fmt.Println("Clearing down existing test orgs/spaces...")
	// Infinite memory quota OO
	setup = workflowhelpers.NewRunawayAppTestSuiteSetup(cfg)

	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		if cfg.UseExistingOrganization {
			cleanupApps()
		} else {
			cleanupOrg()
		}
	})
	fmt.Println("Clearing down existing test orgs/spaces... Complete")
}

func cleanupOrg() {
	orgs := GetTestOrgs(cfg)
	for _, org := range orgs {
		orgName, _, spaceName, _ := GetOrgSpaceNamesAndGuids(cfg, org)
		if spaceName != "" {
			DeleteOrgWithTimeout(orgName, time.Duration(120)*time.Second)
		}
	}
}

func cleanupApps() {
	org := cfg.ExistingOrganization

}

func updateOrgQuota(name string, appCount int, timeout time.Duration) {
	args := []string{"update-org-quota", name}
	args = append(args, "-r", strconv.Itoa(appCount*2))
	args = append(args, "-s", strconv.Itoa(appCount*2))
	args = append(args, "-m", strconv.Itoa(appCount*256)+"MB")
	args = append(args, "--reserved-route-ports", "-1")
	updateOrgQuota := cf.Cf(args...).Wait(timeout)
	Expect(updateOrgQuota).To(gexec.Exit(0), "unable update org quota: "+string(updateOrgQuota.Out.Contents()[:]))
}
