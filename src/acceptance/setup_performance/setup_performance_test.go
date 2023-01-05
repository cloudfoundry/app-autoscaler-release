package peformance_setup_test

import (
	"acceptance/helpers"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Prepare test apps based on benchmark inputs", func() {
	var (
		appName          string
		runningAppsCount int32
		pendingApps      sync.Map
		errors           sync.Map
	)

	AfterEach(func() {
		pendingApps.Range(func(k, v interface{}) bool {
			fmt.Printf("pending app: %s \n", k)
			return true
		})

		errors.Range(func(appName, err interface{}) bool {
			fmt.Printf("errors by app: %s: %s \n", appName, err.(error).Error())
			return true
		})
	})
	BeforeEach(func() {
		workerCount := cfg.Performance.SetupWorkers
		appsChan := make(chan string)

		wg := sync.WaitGroup{}
		// ToDo: Why starting workers first than putting the appname to appChannel.
		// it should be .. first appchanel then start workers
		fmt.Println(fmt.Sprintf("\nStarting %d workers...", cfg.Performance.SetupWorkers))
		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			go worker(appsChan, &runningAppsCount, &pendingApps, &errors, &wg)
		}

		fmt.Println(fmt.Sprintf("Deploying %d apps", cfg.Performance.AppCount))
		for i := 0; i < cfg.Performance.AppCount; i++ {
			appName = fmt.Sprintf("node-custom-metric-benchmark-%d", i)
			pendingApps.Store(appName, 1)
			appsChan <- appName
		}

		close(appsChan)
		wg.Wait()

		fmt.Printf("Total Running apps: %d/%dn", atomic.LoadInt32(&runningAppsCount), cfg.Performance.AppCount)

	})

	Context("when scaling by custom metrics", func() {
		It("should scale out and scale in", func() {
			//Eventually(func() int32 { return atomic.LoadInt32(&runningAppsCount) }, 3*time.Minute, 5*time.Second).Should(BeEquivalentTo(cfg.Performance.AppCount))
		})
	})
})

func worker(appsChan chan string, runningApps *int32, pendingApps *sync.Map, errors *sync.Map, wg *sync.WaitGroup) {
	// Decreasing internal counter for wait-group as soon as goroutine finishes
	defer wg.Done()
	defer GinkgoRecover()
	for appName := range appsChan {
		err := helpers.CreateTestAppFromDropletByName(cfg, nodeAppDropletPath, appName, 1)
		if err != nil {
			errors.Store(appName, err)
			continue
		}
		policy := helpers.GenerateDynamicScaleOutAndInPolicy(1, 2, "test_metric", 500, 500)
		appGUID, err := helpers.GetAppGuid(cfg, appName)
		if err != nil {
			errors.Store(appName, err)
			continue
		}
		_, err = helpers.CreatePolicyWithErr(cfg, appName, appGUID, policy)
		if err != nil {
			errors.Store(appName, err)
			continue
		}
		helpers.CreateCustomMetricCred(cfg, appName, appGUID)
		err = helpers.StartAppWithErr(appName, cfg.CfPushTimeoutDuration())
		if err != nil {
			errors.Store(appName, err)
			continue
		}
		atomic.AddInt32(runningApps, 1)
		pendingApps.Delete(appName)
		fmt.Printf("Running apps: %d/%d - %s\n", atomic.LoadInt32(runningApps), cfg.Performance.AppCount, appName)
	}
}
