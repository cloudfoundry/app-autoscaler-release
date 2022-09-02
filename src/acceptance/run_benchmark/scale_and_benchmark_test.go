package run_benchmark

import (
	"acceptance/helpers"
	"fmt"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega/gmeasure"
	"math"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scale in and out (eg: 30%) percentage of apps", func() {
	var (
		appsToScaleCount int
		samplingConfig gmeasure.SamplingConfig
		experiment *gmeasure.Experiment

	)

	BeforeEach(func() {
		appsToScaleCount = int(math.RoundToEven(float64(cfg.BenchmarkAppCount * cfg.BenchmarkPercentageToScale / 100)))
		ginkgo.GinkgoWriter.Printf("\nScaling %d app: \n", appsToScaleCount)
		samplingConfig = gmeasure.SamplingConfig{
			N: appsToScaleCount,
			NumParallel: appsToScaleCount,
			Duration: 20 * time.Minute,
		}
		experiment = gmeasure.NewExperiment("Scaling Benchmark")
	})

	//	//we get the median repagination duration from the experiment we just ran
	//	repaginationStats := experiment.GetStats("repagination")
	//	medianDuration := repaginationStats.DurationFor(gmeasure.StatMedian)

	//	//and assert that it hasn't changed much from ~100ms
	//	Expect(medianDuration).To(BeNumerically("~", 100*time.Millisecond, 50*time.Millisecond))
	//})

	Context("when scaling by custom metrics",func() {
		It("should scale in", Serial, Label("measurement"), func() {
			AddReportEntry(experiment.Name, experiment)


			var scaledOut chan string
			scaledOut = make(chan string,appsToScaleCount )

			var scaledIn chan string
			scaledIn = make(chan string,appsToScaleCount )

			experiment.Sample(func(i int) {
				appName := fmt.Sprintf("node-custom-metric-benchmark-%d", i+1)
				appGUID := helpers.GetAppGuid(cfg, appName)

				experiment.MeasureDuration("scale-out", func() {
					scaleOut := func() int {
						helpers.SendMetric(cfg, appName, 550)
						return helpers.RunningInstances(appGUID, 5*time.Second)
					}
					Eventually(scaleOut, 5*time.Minute, 15*time.Second).Should(Equal(2))
					scaledOut <- appName
				})

				Eventually(func() int {return len(scaledOut)},10* time.Minute, 5 * time.Second).Should(Equal(appsToScaleCount))

				 experiment.MeasureDuration("scale-in", func() {
					  scaleIn := func() int {
						   helpers.SendMetric(cfg, appName, 100)
						   //TODO: Change to autoscaler-app-history event instead of cf running instances.
						   return helpers.RunningInstances(appGUID, 5*time.Second)
					  }
					 Eventually(scaleIn, 5*time.Minute, 15*time.Second).Should(Equal(1))
					 scaledIn <- appName
				 })
			}, samplingConfig)

			Eventually(func() int {return len(scaledIn)},10* time.Minute, 5 * time.Second).Should(Equal(appsToScaleCount))
		})

	})


})
