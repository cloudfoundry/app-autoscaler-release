package pre_upgrade_test

import (
	"acceptance/helpers"
	"fmt"
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
		workerCount := cfg.Performance.SetupWorkers
		appsChan := make(chan string)

		By(fmt.Sprintf("Deploying %d apps", cfg.Performance.AppCount))
		wg := sync.WaitGroup{}
		//wg.Add(cfg.BenchmarkAppCount)

		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go worker(appsChan, &runningApps, &wg)
		}

		for i := 0; i < cfg.Performance.AppCount; i++ {
			appName = fmt.Sprintf("node-custom-metric-benchmark-%d", i)
			appsChan <- appName
		}

		close(appsChan)
		wg.Wait()
	})

	Context("when scaling by custom metrics", func() {
		It("should scale out and scale in", func() {
			Eventually(func() int32 { return atomic.LoadInt32(&runningApps) }, 3*time.Minute, 5*time.Second).Should(BeEquivalentTo(cfg.Performance.AppCount))
		})
	})
})

func worker(appsChan chan string, runningApps *int32, wg *sync.WaitGroup) {
	// Decreasing internal counter for wait-group as soon as goroutine finishes
	defer wg.Done()
	defer GinkgoRecover()
	for appName := range appsChan {
		helpers.CreateTestAppFromDropletByName(cfg, nodeAppDropletPath, appName, 1)
		policy := helpers.GenerateDynamicScaleOutAndInPolicy(1, 2, "test_metric", 500, 500)
		appGUID := helpers.GetAppGuid(cfg, appName)
		_ = helpers.CreatePolicy(cfg, appName, appGUID, policy)
		helpers.CreateCustomMetricCred(cfg, appName, appGUID)
		helpers.StartApp(appName, cfg.CfPushTimeoutDuration())
		atomic.AddInt32(runningApps, 1)
		GinkgoWriter.Printf("\nRunning apps: %d/%d \n", atomic.LoadInt32(runningApps), cfg.Performance.AppCount)
	}
}
