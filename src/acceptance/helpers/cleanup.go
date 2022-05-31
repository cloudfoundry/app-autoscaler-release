package helpers

import (
	"code.cloudfoundry.org/app-autoscaler/src/acceptance/config"
	"fmt"

	"github.com/onsi/ginkgo/v2"

	"github.com/KevinJCross/cf-test-helpers/v2/cf"
	"github.com/KevinJCross/cf-test-helpers/v2/workflowhelpers"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func Cleanup(cfg *config.Config, wfh *workflowhelpers.ReproducibleTestSuiteSetup) {
	ginkgo.By("Clearing down existing test orgs/spaces...")
	workflowhelpers.AsUser(wfh.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		orgs := GetTestOrgs(cfg)

		for _, org := range orgs {
			orgName, orgGuid, spaceName, spaceGuid := GetOrgSpaceNamesAndGuids(cfg, org)
			if spaceName != "" {
				target := cf.Cf("target", "-o", orgName, "-s", spaceName).Wait(cfg.DefaultTimeoutDuration())
				Expect(target).To(Exit(0), fmt.Sprintf("failed to target %s and %s", orgName, spaceName))

				apps := GetApps(cfg, orgGuid, spaceGuid, "autoscaler-")
				DeleteApps(cfg, apps, 0)

				services := GetServices(cfg, orgGuid, spaceGuid, "autoscaler-")
				DeleteServices(cfg, services)
			}

			DeleteOrg(cfg, org)
		}
	})
	ginkgo.By("Clearing down existing test orgs/spaces... Complete")
}
