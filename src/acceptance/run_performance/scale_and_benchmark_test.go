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

var _ = Describe("Scale in and out (eg: 30%) percentage of apps", func() {
	var (
		appsToScaleCount   int
		samplingConfig     gmeasure.SamplingConfig
		experiment         *gmeasure.Experiment
		doneAppsCount      int32
		scaledOutAppsCount int32
	)

	BeforeEach(func() {
		appsToScaleCount = int(math.RoundToEven(float64(cfg.Performance.AppCount * cfg.Performance.PercentageToScale / 100)))
		GinkgoWriter.Printf("\nScaling %d app: \n", appsToScaleCount)
		samplingConfig = gmeasure.SamplingConfig{
			N:           appsToScaleCount,
			NumParallel: appsToScaleCount,
			Duration:    20 * time.Minute,
		}
		experiment = gmeasure.NewExperiment("Scaling Benchmark")
	})

	Context("when scaling by custom metrics", func() {
		It("should scale in", Serial, Label("measurement"), func() {
			AddReportEntry(experiment.Name, experiment)
			//experimentWg := sync.WaitGroup{}
			//	experimentWg.Add(appsToScaleCount)

			experiment.Sample(func(i int) {
				defer GinkgoRecover()
				appName := fmt.Sprintf("node-custom-metric-benchmark-%d", i+1)
				appGUID := helpers.GetAppGuid(cfg, appName)
				pollTime := 10 * time.Second

				wg := sync.WaitGroup{}
				wg.Add(1)
				experiment.MeasureDuration("scale-out", func() {
					scaleOut := func() int {
						helpers.SendMetric(cfg, appName, 550)
						return helpers.RunningInstances(appGUID, 5*time.Second)
					}
					Eventually(scaleOut).WithPolling(pollTime).WithTimeout(5*time.Minute).Should(Equal(2),
						fmt.Sprintf("Failed to scale out app: %s", appName))
					wg.Done()
				})
				wg.Wait()

				atomic.AddInt32(&scaledOutAppsCount, 1)
				fmt.Println(fmt.Sprintf("Scaled-Out apps: %d/%d", atomic.LoadInt32(&scaledOutAppsCount), appsToScaleCount))

				wg = sync.WaitGroup{}
				wg.Add(1)
				experiment.MeasureDuration("scale-in", func() {
					scaleIn := func() int {
						helpers.SendMetric(cfg, appName, 100)
						return helpers.RunningInstances(appGUID, 5*time.Second)
					}
					Eventually(scaleIn).WithPolling(pollTime).WithTimeout(5*time.Minute).Should(Equal(1),
						fmt.Sprintf("Failed to scale in app: %s", appName))
					wg.Done()
				})
				wg.Wait()

				atomic.AddInt32(&doneAppsCount, 1)
				fmt.Println(fmt.Sprintf("Scaled-in apps: %d/%d", atomic.LoadInt32(&doneAppsCount), appsToScaleCount))
				//experimentWg.Done()

			}, samplingConfig)
			//experimentWg.Wait()
			Eventually(func() int32 { return atomic.LoadInt32(&doneAppsCount) }, 10*time.Minute, 5*time.Second).Should(BeEquivalentTo(appsToScaleCount))
			checkMedianDurationFor(experiment, "scale-out")
			checkMedianDurationFor(experiment, "scale-in")
		})
	})
})

func checkMedianDurationFor(experiment *gmeasure.Experiment, statName string) {
	scaleOutStats := experiment.GetStats(statName)
	medianScaleOutDuration := scaleOutStats.DurationFor(gmeasure.StatMedian)
	fmt.Println(fmt.Sprintf("%d duration:", medianScaleOutDuration))
}
