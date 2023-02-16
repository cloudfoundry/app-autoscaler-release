package run_performance_test

import (
	"acceptance/helpers"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
)

const pollTime = 10 * time.Second
const timeOutDuration = 20 * time.Minute
const desiredScalingTime = 2 * time.Hour

var _ = Describe("Scale in and out (eg: 30%) percentage of apps", Ordered, func() {
	var (
		appsToScaleCount       int
		percentageToScale      int
		appCount               int
		samplingConfig         gmeasure.SamplingConfig
		experiment             = gmeasure.NewExperiment("Scaling Benchmark")
		scaledInAppsCount      atomic.Int32
		scaledOutAppsCount     atomic.Int32
		startedApps            []helpers.AppInfo
		actualAppsToScaleCount int
		pendingScaleOuts       sync.Map
		pendingScaleIns        sync.Map
		scaleOutApps           sync.Map
		scaleInApps            sync.Map
	)
	AfterEach(func() {
		fmt.Println("\n\nSummary...")

		fmt.Println("\nSuccessful Scale-Out...")
		scaleOutApps.Range(func(appName, appGuid interface{}) bool {
			fmt.Printf("scale-out successful: %s: %s \n", appName, appGuid)
			return true
		})
		fmt.Println("\nSuccessful Scale-In")
		scaleInApps.Range(func(appName, appGuid interface{}) bool {
			fmt.Printf("scale-In successful: %s: %s \n", appName, appGuid)
			return true
		})

		fmt.Println("\nScale-Out Errors")
		pendingScaleOuts.Range(func(appName, appGuid interface{}) bool {
			fmt.Printf("scale-out app error: %s: %s \n", appName, appGuid)
			return true
		})

		fmt.Println("\nScale-In Errors")
		pendingScaleIns.Range(func(appName, appGuid interface{}) bool {
			fmt.Printf("scale-in app error: %s: %s \n", appName, appGuid)
			return true
		})
	})

	BeforeEach(func() {
		orgGuid := helpers.GetOrgGuid(cfg, cfg.ExistingOrganization)
		spaceGuid := helpers.GetSpaceGuid(cfg, orgGuid)
		startedApps = helpers.GetAllStartedApp(cfg, orgGuid, spaceGuid, "node-custom-metric-benchmark")

		percentageToScale, appCount = cfg.Performance.PercentageToScale, cfg.Performance.AppCount
		if percentageToScale < 0 || percentageToScale > 100 {
			err := fmt.Errorf(
				"Given scaling percentage not in [0, 100] which does not make sense: percentageToScale = %d",
				percentageToScale)
			Fail(err.Error())
		}
		appsToScaleCount = appCount * percentageToScale / 100
		Expect(appsToScaleCount).To(BeNumerically(">", 0),
			fmt.Sprintf("%d percent of %d must round up to 1 or more app(s)", percentageToScale, appCount))

	})

	Context("when scaling out by custom metrics", func() {
		JustBeforeEach(func() {
			fmt.Printf("=====running JustBeforeEach=======")
			// Now calculate appsToScaleCount based on the actual startedApps
			actualAppsToScaleCount = len(startedApps) * percentageToScale / 100

			//fmt.Printf("Debug-Apps: started apps... %v\n\n%d", startedApps, len(startedApps))
			fmt.Printf("\nDesired Scaling %d apps \n", appsToScaleCount)
			fmt.Printf("Actual Scaling %d apps (based on successful apps pushed) \n\n", actualAppsToScaleCount)

			samplingConfig = gmeasure.SamplingConfig{
				N:           actualAppsToScaleCount,
				NumParallel: 50, // number of parallel node/process  to execute at a time e.g. 50 scaleout will run on 50 nodes
				Duration:    5 * time.Hour,
			}
		})
		It("should scale out", Label("measurement"), func() {
			AddReportEntry(experiment.Name, experiment)
			wg := sync.WaitGroup{}
			wg.Add(samplingConfig.N)

			experiment.Sample(func(workerIndex int) {
				defer GinkgoRecover()
				defer wg.Done()

				tasksWG := sync.WaitGroup{}
				tasksWG.Add(1)
				appName := startedApps[workerIndex].Name
				appGUID := startedApps[workerIndex].Guid

				pendingScaleOuts.Store(appName, appGUID)
				experiment.MeasureDuration("scale-out",
					scaleOutApp(appName, appGUID, &scaleOutApps, &pendingScaleOuts, &scaledOutAppsCount,
						actualAppsToScaleCount, workerIndex, &tasksWG))

				tasksWG.Wait()

			}, samplingConfig)

			fmt.Printf("Waiting for scale-out workers to finish scaling...\n\n")
			wg.Wait()
			fmt.Printf("\nTotal scale out apps: %d/%d\n", scaledOutAppsCount.Load(), actualAppsToScaleCount)

			fmt.Println("\nScale-Out Errors")
			pendingScaleOuts.Range(func(appName, appGuid interface{}) bool {
				fmt.Printf("scale-out app error: %s: %s \n", appName, appGuid)
				return true
			})

		})

	})
	Context("scale-out results", func() {
		It("wait for scale-out results", func() {
			Eventually(func() int32 {
				count := scaledOutAppsCount.Load()
				fmt.Printf("current scaledOutAppsCount %d\n", count)
				return count
			}).WithPolling(10 * time.Second).
				WithTimeout(desiredScalingTime).
				Should(BeEquivalentTo(actualAppsToScaleCount))
		})
	})
	Context("when scaling In by custom metrics", func() {
		JustBeforeEach(func() {
			fmt.Printf("\nsuccessfull scale-out count: %d\n", lenOfSyncMap(&scaleOutApps))
		})
		It("should scale in", Label("measurement"), func() {
			AddReportEntry(experiment.Name, experiment)
			wg := sync.WaitGroup{}
			wg.Add(samplingConfig.N)

			experiment.Sample(func(workerIndex int) {
				defer GinkgoRecover()
				defer wg.Done()

				value, ok := scaleOutApps.Load(workerIndex)
				if !ok {
					fmt.Printf("\nunable to find scaled app at worker %d\n", workerIndex)
					return
				}
				// cast to struct in better way
				scaledOutApps := value.(helpers.AppInfo)
				appName := scaledOutApps.Name
				appGUID := scaledOutApps.Guid
				taskWG := sync.WaitGroup{}
				taskWG.Add(1)

				pendingScaleIns.Store(appName, appGUID)
				experiment.MeasureDuration("scale-in",
					scaleInApp(appName, appGUID, &scaleInApps, &pendingScaleIns,
						&scaledInAppsCount, actualAppsToScaleCount, workerIndex, &taskWG))
				taskWG.Wait()

			}, samplingConfig)

			fmt.Printf("Waiting for scale-In workers to finish scaling...\n\n")
			wg.Wait()
			fmt.Printf("\nTotal scale In apps: %d/%d\n", scaledInAppsCount.Load(), actualAppsToScaleCount)

		})
	})
	Context("scale-In results", func() {
		It("wait for scale-in results", func() {

			Eventually(func() int32 {
				count := scaledInAppsCount.Load()
				fmt.Printf("current scaledInAppsCount %d\n", count)
				return count
			}).WithPolling(10 * time.Second).
				WithTimeout(desiredScalingTime).
				Should(BeEquivalentTo(actualAppsToScaleCount))

			checkMedianDurationFor(experiment, "scale-out")
			checkMedianDurationFor(experiment, "scale-in")
		})
	})
})

func scaleOutApp(appName string, appGUID string, scaleOutApps *sync.Map,
	pendingScaleOuts *sync.Map, scaledOutAppsCount *atomic.Int32, actualAppsToScaleCount int, workerIndex int, wg *sync.WaitGroup) func() {
	return func() {
		scaleOut := func() (int, error) {
			cmdOutput := helpers.SendMetricWithTimeout(cfg, appName, 550, 10*time.Minute)
			fmt.Printf("worker %d -  %s with App %s %s\n\n", workerIndex, cmdOutput, appName, appGUID)
			instances, err := helpers.RunningInstances(appGUID, 10*time.Minute)
			if err != nil {
				fmt.Printf("		error running instances for app %s %s\n", appName, appGUID)
				return 0, err
			}
			fmt.Printf("		worker %d - running instances for app %s %s - %d\n", workerIndex, appName, appGUID, instances)
			return instances, nil
		}
		fmt.Printf("		worker %d - scale-out starts for app %s with AppGuid %s\n", workerIndex, appName, appGUID)
		Eventually(scaleOut).
			WithPolling(pollTime).
			WithTimeout(timeOutDuration).
			Should(Equal(2),
				fmt.Sprintf("Failed to scale out app: %s", appName))
		fmt.Printf("		debug - worker %d - finished - trying to add 1 to scaleOutAppCount %s %s\n", workerIndex, appName, appGUID)
		scaledOutAppsCount.Add(1)
		fmt.Printf("		debug - worker %d - Scaled-Out apps: %d/%d – size of pendinScaleOuts: %d\n", workerIndex,
			scaledOutAppsCount.Load(), actualAppsToScaleCount, lenOfSyncMap(pendingScaleOuts))
		scaleOutApps.Store(workerIndex, helpers.AppInfo{Name: appName, Guid: appGUID})
		pendingScaleOuts.Delete(appName)

		defer wg.Done()
	}
}

func scaleInApp(appName string, appGUID string, scaleInApps *sync.Map, pendingScaleIns *sync.Map,
	scaledInAppsCount *atomic.Int32, actualAppsToScaleCount int, workerIndex int, wg *sync.WaitGroup) func() {
	return func() {
		scaleIn := func() (int, error) {
			cmdOutput := helpers.SendMetricWithTimeout(cfg, appName, 100, 10*time.Minute)
			fmt.Printf("worker %d -  %s with App %s %s\n\n", workerIndex, cmdOutput, appName, appGUID)
			return helpers.RunningInstances(appGUID, 10*time.Minute)
		}
		fmt.Printf("		worker %d - scale-in starts for app %s with AppGuid %s\n", workerIndex, appName, appGUID)
		Eventually(scaleIn).
			WithPolling(pollTime).
			WithTimeout(timeOutDuration).
			Should(Equal(1),
				fmt.Sprintf("Failed to scale in app: %s", appName))

		fmt.Printf("debug - worker %d - finished - trying to add 1 to scaledInAppsCount %s %s\n", workerIndex, appName, appGUID)
		scaledInAppsCount.Add(1)
		fmt.Printf("debug - worker %d - Scaled-In apps: %d/%d – size of pendinScaleOuts: %d\n",
			workerIndex, scaledInAppsCount.Load(), actualAppsToScaleCount, lenOfSyncMap(pendingScaleIns))
		scaleInApps.Store(workerIndex, helpers.AppInfo{Name: appName, Guid: appGUID})
		pendingScaleIns.Delete(appName)

		defer wg.Done()
	}
}

func checkMedianDurationFor(experiment *gmeasure.Experiment, statName string) {
	stats := experiment.GetStats(statName)
	medianDuration := stats.DurationFor(gmeasure.StatMedian)
	fmt.Printf("\nMedian duration for %s: %d", statName, medianDuration)
}

func lenOfSyncMap(m *sync.Map) int32 {
	var counter atomic.Int32
	m.Range(func(_ any, _ any) bool {
		counter.Add(1)
		return true
	})
	return counter.Load()
}
