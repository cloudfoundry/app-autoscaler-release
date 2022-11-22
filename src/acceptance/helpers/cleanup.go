package helpers

import (
	"acceptance/config"

	"github.com/onsi/ginkgo/v2"

	"github.com/KevinJCross/cf-test-helpers/v2/workflowhelpers"
)

func CleanupOrgs(cfg *config.Config, wfh *workflowhelpers.ReproducibleTestSuiteSetup) {
	ginkgo.By("Clearing down existing test orgs/spaces...")
	workflowhelpers.AsUser(wfh.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		orgs := GetTestOrgs(cfg)
		for _, org := range orgs {
			DeleteOrg(cfg, org)
		}
	})
	ginkgo.By("Clearing down existing test orgs/spaces... Complete")
}
