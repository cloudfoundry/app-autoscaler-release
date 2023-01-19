package run_performance_test

import (
	"acceptance/helpers"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
)

const pollTime = 10 * time.Second
const desiredScalingTime = 300 * time.Minute

var _ = Describe("Scale in and out (eg: 30%) percentage of apps", func() {
	var (
		appsToScaleCount       int
		percentageToScale      int
		appCount               int
		samplingConfig         gmeasure.SamplingConfig
		experiment             *gmeasure.Experiment
		scaledInAppsCount      int32
		scaledOutAppsCount     int32
		startedApps            []helpers.AppInfo
		actualAppsToScaleCount int
		pendingScaleOuts       sync.Map
		pendingScaleIns        sync.Map
		scaleOutApps           sync.Map
		scaleInApps            sync.Map
	)
	AfterEach(func() {
		fmt.Println("==Errors===")
		scaleOutApps.Range(func(appName, appGuid interface{}) bool {
			//fmt.Printf("scale-out successful: %s: %s \n", appName, appGuid)
			return true
		})

		pendingScaleOuts.Range(func(appName, appGuid interface{}) bool {
			fmt.Printf("scale-out app error: %s: %s \n", appName, appGuid)
			return true
		})

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
		appsToScaleCount = int(math.RoundToEven(float64(appCount * percentageToScale / 100)))
		Expect(appsToScaleCount).To(BeNumerically(">", 0),
			fmt.Sprintf("%d percent of %d must round up to 1 or more app(s)", percentageToScale, appCount))

		// Now calculate appsToScaleCount based on the actual startedApps
		actualAppsToScaleCount = int(math.RoundToEven(float64(len(startedApps) * percentageToScale / 100)))

		fmt.Printf("Debug-Apps ready to scale %+v\n", startedApps)
		fmt.Printf("\nDesired Scaling %d apps \n", appsToScaleCount)
		fmt.Printf("Actual Scaling %d apps (based on sucessuful apps pushed) \n\n", actualAppsToScaleCount)

		samplingConfig = gmeasure.SamplingConfig{
			N:           actualAppsToScaleCount,
			NumParallel: 100, // number of sample to execute at a time
			Duration:    300 * time.Minute,
		}
		experiment = gmeasure.NewExperiment("Scaling Benchmark")
	})

	Context("when scaling by custom metrics", func() {
		It("should scale in", Serial, Label("measurement"), func() {
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(i int) {
				defer GinkgoRecover()
				appName := startedApps[i].Name
				appGUID := startedApps[i].Guid
				//fmt.Printf("Debug-scaling candidate %+v\n", startedApps[i])
				wg := sync.WaitGroup{}
				wg.Add(1)
				pendingScaleOuts.Store(appName, appGUID)
				experiment.MeasureDuration("scale-out", scaleOutApp(appName, appGUID, &scaleOutApps, &pendingScaleOuts, &wg))
				//fmt.Printf("Waiting for scale out to finish scaling...%s\n", appName)
				wg.Wait()

				atomic.AddInt32(&scaledOutAppsCount, 1)
				fmt.Printf("Scaled-Out apps: %d/%d\n", atomic.LoadInt32(&scaledOutAppsCount), actualAppsToScaleCount)

				// Do not perform scale in - if scale out of an app was unsuccessful due to any reason
				_, isValuePresent := pendingScaleOuts.Load(appName)
				if isValuePresent {
					fmt.Printf("skipping scale-in as scale out was unsuccessful for app %s\n", appName)
					return
				}

				// perform scale-in
				wg = sync.WaitGroup{}
				wg.Add(1)
				pendingScaleIns.Store(appName, appGUID)
				experiment.MeasureDuration("scale-in", scaleInApp(appName, appGUID, &scaleInApps, &pendingScaleOuts, &wg))
				//fmt.Printf("Waiting for scale in to finish scaling...%s\n", appName)
				wg.Wait()
				atomic.AddInt32(&scaledInAppsCount, 1)
				fmt.Printf("Scaled-in apps: %d/%d\n", atomic.LoadInt32(&scaledInAppsCount), actualAppsToScaleCount)

			}, samplingConfig)

			fmt.Printf("Waiting %s minutes to finish scaling...\n\n", desiredScalingTime)
			// ToDo based on dynamic calculation
			Eventually(func() int32 { return atomic.LoadInt32(&scaledInAppsCount) }, desiredScalingTime, 10*time.Second).Should(BeNumerically(">", 500))
			checkMedianDurationFor(experiment, "scale-out")
			checkMedianDurationFor(experiment, "scale-in")
		})
	})
})

func scaleInApp(appName string, appGUID string, scaleInApps *sync.Map, pendingScaleIns *sync.Map, wg *sync.WaitGroup) func() {
	return func() {
		scaleIn := func() (int, error) {
			helpers.SendMetric(cfg, appName, 100)
			return helpers.RunningInstances(appGUID, 20*time.Second)
		}
		Eventually(scaleIn).WithPolling(pollTime).WithTimeout(5*time.Minute).Should(Equal(1),
			fmt.Sprintf("Failed to scale in app: %s", appName))
		scaleInApps.Store(appName, appGUID)
		pendingScaleIns.Delete(appName)
		wg.Done()
	}
}

func scaleOutApp(appName string, appGUID string, scaleOutApps *sync.Map, pendingScaleOuts *sync.Map, wg *sync.WaitGroup) func() {
	return func() {
		scaleOut := func() (int, error) {
			helpers.SendMetric(cfg, appName, 550)
			return helpers.RunningInstances(appGUID, 20*time.Second)
		}
		Eventually(scaleOut).WithPolling(pollTime).WithTimeout(5*time.Minute).Should(Equal(2),
			fmt.Sprintf("Failed to scale out app: %s", appName))
		scaleOutApps.Store(appName, appGUID)
		pendingScaleOuts.Delete(appName)

		defer wg.Done()
	}
}

func checkMedianDurationFor(experiment *gmeasure.Experiment, statName string) {
	stats := experiment.GetStats(statName)
	medianDuration := stats.DurationFor(gmeasure.StatMedian)
	fmt.Printf("\nMedian duration for %s: %d", statName, medianDuration)
}
