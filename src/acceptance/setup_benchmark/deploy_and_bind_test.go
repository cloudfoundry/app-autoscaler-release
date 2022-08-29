package pre_upgrade_test

import (
	"acceptance/helpers"
	"fmt"
	"github.com/onsi/ginkgo/v2"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Prepare test apps based on benchmark inputs", func() {
	var (
		appName string
		orgGuid string
		spaceGuid string
	)

	BeforeEach(func() {
		ginkgo.GinkgoWriter.Printf("\nDeploying %d app: \n", cfg.AppCount)

		for i := 1; i <= cfg.AppCount; i++ {
			appName = fmt.Sprintf("node-custom-metric-benchmark-%d", i)

			go func(appName string) {
				defer GinkgoRecover()
				helpers.CreateTestAppByName(*cfg, appName, 1)
				policy := helpers.GenerateDynamicScaleOutAndInPolicy(1, 2, "test_metric", 500, 500)
				appGUID := helpers.GetAppGuid(cfg, appName)
				_ = helpers.CreatePolicy(cfg, appName, appGUID, policy)
				helpers.CreateCustomMetricCred(cfg, appName, appGUID)
			}(appName)
		}

		_, orgGuid, _, spaceGuid = helpers.GetOrgSpaceNamesAndGuids(cfg, helpers.GetTestOrgs(cfg)[0])

		appsDeployed := func() int {
			var apps []string

			apps = helpers.GetApps(cfg, orgGuid, spaceGuid, "node-custom-metric-benchmark")
			ginkgo.GinkgoWriter.Printf("\nGot apps: %s\n", apps)

			return len(apps)
		}

		Eventually(appsDeployed, 5*time.Minute, 5*time.Second).Should(Equal(cfg.AppCount))
	})

	Context("when scaling by custom metrics", func() {
		It("should scale out and scale in", func() {
			for i := 1; i <= cfg.AppCount; i++ {
				appName = fmt.Sprintf("node-custom-metric-benchmark-%d", i)
				go func(appName string) {
					defer GinkgoRecover()
					helpers.StartApp(appName, cfg.CfPushTimeoutDuration())
				}(appName)
			}

			appsRunning := func() int {
				var apps []string

				apps = helpers.GetRunningApps(cfg, orgGuid, spaceGuid)
				ginkgo.GinkgoWriter.Printf("\nGot running apps: %s\n", apps)

				return len(apps)
			}

			Eventually(appsRunning, 2*time.Minute, 5*time.Second).Should(Equal(cfg.AppCount))
		})
	})
})
