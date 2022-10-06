package run_performance_test

import (
	"acceptance/helpers"
	"fmt"
	"math"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
)

var _ = Describe("Scale in and out (eg: 30%) percentage of apps", func() {
	var (
		appsToScaleCount int
		samplingConfig   gmeasure.SamplingConfig
		experiment       *gmeasure.Experiment
	)

	BeforeEach(func() {
		appsToScaleCount = int(math.RoundToEven(float64(cfg.BenchmarkAppCount * cfg.BenchmarkPercentageToScale / 100)))
		GinkgoWriter.Printf("\nScaling %d app: \n", appsToScaleCount)
		samplingConfig = gmeasure.SamplingConfig{
			N:           appsToScaleCount,
			NumParallel: appsToScaleCount,
			Duration:    20 * time.Minute,
		}
		experiment = gmeasure.NewExperiment("Scaling Benchmark")
	})

	//	//we get the median repagination duration from the experiment we just ran
	//	repaginationStats := experiment.GetStats("repagination")
	//	medianDuration := repaginationStats.DurationFor(gmeasure.StatMedian)

	//	//and assert that it hasn't changed much from ~100ms
	//	Expect(medianDuration).To(BeNumerically("~", 100*time.Millisecond, 50*time.Millisecond))
	//})

	Context("when scaling by custom metrics", func() {
		It("should scale in", Serial, Label("measurement"), func() {
			AddReportEntry(experiment.Name, experiment)
			experimentWg := sync.WaitGroup{}
			experimentWg.Add(appsToScaleCount)

			experiment.Sample(func(i int) {
				defer GinkgoRecover()
				appName := fmt.Sprintf("node-custom-metric-benchmark-%d", i+1)
				appGUID := helpers.GetAppGuid(cfg, appName)
				pollTime := 10 * time.Second

				wg := sync.WaitGroup{}
				wg.Add(1)
				experiment.MeasureDuration("scale-out", func() {
					Eventually(func() int {
						helpers.SendMetric(cfg, appName, 550)
						return helpers.RunningInstances(appGUID, 5*time.Second)
					}).WithPolling(pollTime).WithTimeout(5 * time.Minute).Should(Equal(2))
					wg.Done()
				})
				wg.Wait()

				wg = sync.WaitGroup{}
				wg.Add(1)
				experiment.MeasureDuration("scale-in", func() {
					scaleIn := func() int {
						helpers.SendMetric(cfg, appName, 100)
						//TODO: Change to autoscaler-app-history event instead of cf running instances.
						return helpers.RunningInstances(appGUID, 5*time.Second)
					}
					Eventually(scaleIn).WithPolling(pollTime).WithTimeout(5 * time.Minute).Should(Equal(1))
					wg.Done()
				})
				wg.Wait()

				experimentWg.Done()
			}, samplingConfig)
			experimentWg.Wait()
		})

	})

})
