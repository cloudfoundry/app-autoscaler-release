package pre_upgrade_test

import (
	"acceptance/helpers"
	"fmt"
	"github.com/onsi/ginkgo/v2"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Prepare test apps based on benchmark inputs", func() {
	var (
		appName     string
		runningApps int32
	)

	BeforeEach(func() {
		wg := sync.WaitGroup{}
		wg.Add(cfg.BenchmarkAppCount)
		By(fmt.Sprintf("Deploying %d apps", cfg.BenchmarkAppCount))

		for i := 0; i < cfg.BenchmarkAppCount; i++ {
			appName = fmt.Sprintf("node-custom-metric-benchmark-%d", i)

			go func(appName string) {
				defer GinkgoRecover()
				helpers.CreateTestAppByName(*cfg, appName, 1)
				policy := helpers.GenerateDynamicScaleOutAndInPolicy(1, 2, "test_metric", 500, 500)
				appGUID := helpers.GetAppGuid(cfg, appName)
				_ = helpers.CreatePolicy(cfg, appName, appGUID, policy)
				helpers.CreateCustomMetricCred(cfg, appName, appGUID)
				helpers.StartApp(appName, cfg.CfPushTimeoutDuration())
				wg.Done()
				atomic.AddInt32(&runningApps, 1)
				ginkgo.GinkgoWriter.Printf("\nRunning apps: %d/%d \n", atomic.LoadInt32(&runningApps), cfg.BenchmarkAppCount)
			}(appName)
		}
		wg.Wait()
	})

	Context("when scaling by custom metrics", func() {
		It("should scale out and scale in", func() {
			Eventually(func() int32 { return atomic.LoadInt32(&runningApps) }, 3*time.Minute, 5*time.Second).Should(BeEquivalentTo(cfg.BenchmarkAppCount))
		})
	})
})
