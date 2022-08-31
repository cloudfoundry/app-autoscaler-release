package pre_upgrade_test

import (
	"acceptance/helpers"
	"fmt"
	"github.com/onsi/ginkgo/v2"
	"math"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scale in and out (eg: 30%) percentage of apps", func() {
	var (
		appName string
		scaleAppDurationChan chan time.Duration
		appsToScaleCount int
	)

	BeforeEach(func() {
		appsToScaleCount = int(math.RoundToEven(float64(cfg.BenchmarkAppCount * cfg.BenchmarkPercentageToScale / 100)))
		scaleAppDurationChan = make(chan time.Duration, appsToScaleCount )
		ginkgo.GinkgoWriter.Printf("\nDeploying %d app: \n", cfg.BenchmarkAppCount)

		for i := 1; i <= appsToScaleCount; i++ {
			appName = fmt.Sprintf("node-custom-metric-benchmark-%d", i)

			go func(appName string) {
				defer GinkgoRecover()

				appGUID := helpers.GetAppGuid(cfg, appName)

				start := time.Now()

				By("Scale out to 2 instances")
				scaleOut := func() int {
					helpers.SendMetric(cfg, appName, 550)
					return helpers.RunningInstances(appGUID, 5*time.Second)
				}
				Eventually(scaleOut, 5*time.Minute, 15*time.Second).Should(Equal(2))

				By("Scale in to 1 instances")
				scaleIn := func() int {
					helpers.SendMetric(cfg, appName, 100)
					return helpers.RunningInstances(appGUID, 5*time.Second)
				}
				Eventually(scaleIn, 5*time.Minute, 15*time.Second).Should(Equal(1))

				scaleAppDurationChan <- time.Since(start)

			}(appName)
		}
	})

	Context("when scaling by custom metrics", func() {
		It("should scale out and scale in", func() {
			Eventually(len(scaleAppDurationChan) , 10*time.Minute, 5*time.Second).Should(Equal(appsToScaleCount))
		})
	})
})
