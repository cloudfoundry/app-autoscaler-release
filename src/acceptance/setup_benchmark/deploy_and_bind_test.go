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
		appName         string
		runningAppsChan chan string
	)

	BeforeEach(func() {
		runningAppsChan = make(chan string, cfg.BenchmarkAppCount)
		ginkgo.GinkgoWriter.Printf("\nDeploying %d app: \n", cfg.BenchmarkAppCount)

		for i := 1; i <= cfg.BenchmarkAppCount; i++ {
			appName = fmt.Sprintf("node-custom-metric-benchmark-%d", i)

			go func(appName string) {
				defer GinkgoRecover()

				helpers.CreateTestAppByName(*cfg, appName, 1)
				policy := helpers.GenerateDynamicScaleOutAndInPolicy(1, 2, "test_metric", 500, 500)
				appGUID := helpers.GetAppGuid(cfg, appName)
				_ = helpers.CreatePolicy(cfg, appName, appGUID, policy)
				helpers.CreateCustomMetricCred(cfg, appName, appGUID)
				helpers.StartApp(appName, cfg.CfPushTimeoutDuration())
				runningAppsChan <- appName
				ginkgo.GinkgoWriter.Printf("\nRunning apps: %d/%d \n", len(runningAppsChan), cfg.BenchmarkAppCount)
			}(appName)
		}
	})

	Context("when scaling by custom metrics", func() {
		It("should scale out and scale in", func() {
			Eventually(func() int { return len(runningAppsChan) }, 3*time.Minute, 5*time.Second).Should(Equal(cfg.BenchmarkAppCount))
		})
	})
})
