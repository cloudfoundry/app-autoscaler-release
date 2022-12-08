package peformance_setup_test

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
		appName          string
		runningAppsCount int32
		pendingApps      sync.Map
	)

	AfterEach(func() {
		pendingApps.Range(func(k, v interface{}) bool {
			fmt.Printf("pending app: %s \n", k)
			return true
		})

	})
	BeforeEach(func() {
		workerCount := cfg.Performance.SetupWorkers
		appsChan := make(chan string)

		By(fmt.Sprintf("Deploying %d apps", cfg.Performance.AppCount))
		wg := sync.WaitGroup{}
		//wg.Add(cfg.BenchmarkAppCount)

		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go worker(appsChan, &runningAppsCount, &pendingApps, &wg)
		}

		for i := 0; i < cfg.Performance.AppCount; i++ {
			appName = fmt.Sprintf("node-custom-metric-benchmark-%d", i)
			pendingApps.Store(appName, 1)
			appsChan <- appName
		}

		close(appsChan)
		wg.Wait()
	})

	Context("when scaling by custom metrics", func() {
		It("should scale out and scale in", func() {
			Eventually(func() int32 { return atomic.LoadInt32(&runningAppsCount) }, 3*time.Minute, 5*time.Second).Should(BeEquivalentTo(cfg.Performance.AppCount))
		})
	})
})

func worker(appsChan chan string, runningApps *int32, pendingApps *sync.Map, wg *sync.WaitGroup) {
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
		pendingApps.Delete(appName)
		fmt.Printf("Running apps: %d/%d\n", atomic.LoadInt32(runningApps), cfg.Performance.AppCount)
	}
}
