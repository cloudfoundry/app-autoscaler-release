package app_test

import (
	"acceptance"
	. "acceptance/helpers"
	"fmt"
	"time"

	cfh "github.com/cloudfoundry/cf-test-helpers/v2/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AutoScaler dynamic policy", func() {
	var (
		policy         string
		err            error
		doneChan       chan bool
		doneAcceptChan chan bool
		ticker         *time.Ticker
		maxHeapLimitMb int
	)

	const minimalMemoryUsage = 28 // observed minimal memory usage by the test app

	JustBeforeEach(func() {
		appName = CreateTestApp(cfg, "dynamic-policy", initialInstanceCount)

		appGUID, err = GetAppGuid(cfg, appName)
		Expect(err).NotTo(HaveOccurred())
		StartApp(appName, cfg.CfPushTimeoutDuration())
		instanceName = CreatePolicy(cfg, appName, appGUID, policy)
	})
	BeforeEach(func() {
		maxHeapLimitMb = cfg.NodeMemoryLimit - minimalMemoryUsage
	})

	AfterEach(AppAfterEach)

	Context("when scaling by memoryused", func() {

		Context("There is a scale out and scale in policy", func() {
			var heapToUse int
			BeforeEach(func() {
				heapToUse = min(maxHeapLimitMb, 200)
				expectedAverageUsageAfterScaling := float64(heapToUse)/2 + minimalMemoryUsage
				policy = GenerateDynamicScaleOutAndInPolicy(1, 2, "memoryused", int64(0.9*expectedAverageUsageAfterScaling), int64(heapToUse))
				initialInstanceCount = 1
			})

			It("should scale out and then back in.", Label(acceptance.LabelSmokeTests), func() {
				By(fmt.Sprintf("Use heap %d mb of heap on app", heapToUse))
				CurlAppInstance(cfg, appName, 0, fmt.Sprintf("/memory/%d/5", heapToUse))

				By("wait for scale to 2")
				WaitForNInstancesRunning(appGUID, 2, 5*time.Minute)

				By("Drop memory used by app")
				CurlAppInstance(cfg, appName, 0, "/memory/close")

				By("Wait for scale to minimum instances")
				WaitForNInstancesRunning(appGUID, 1, 5*time.Minute)
			})
		})
	})

	Context("when scaling by memoryutil", func() {

		Context("when memoryutil", func() {
			BeforeEach(func() {
				//current app resident size is 66mb so 66/128mb is 55%
				policy = GenerateDynamicScaleOutAndInPolicy(1, 2, "memoryutil", 58, 63)
				initialInstanceCount = 1
			})

			It("should scale out and back in", func() {
				heapToUse := min(maxHeapLimitMb, int(float64(cfg.NodeMemoryLimit)*0.80))
				By(fmt.Sprintf("use 80%% or %d of memory in app", heapToUse))
				CurlAppInstance(cfg, appName, 0, fmt.Sprintf("/memory/%d/5", heapToUse))

				By("Wait for scale to 2 instances")
				WaitForNInstancesRunning(appGUID, 2, 5*time.Minute)

				By("drop memory used")
				CurlAppInstance(cfg, appName, 0, "/memory/close")

				By("Wait for scale down to 1 instance")
				WaitForNInstancesRunning(appGUID, 1, 5*time.Minute)
			})
		})
	})

	Context("when scaling by responsetime", func() {

		JustBeforeEach(func() {
			doneChan = make(chan bool)
			doneAcceptChan = make(chan bool)
		})

		AfterEach(func() {
			close(doneChan)
			Eventually(doneAcceptChan, 10*time.Second).Should(Receive())
		})

		Context("when responsetime is greater than scaling out threshold", func() {

			BeforeEach(func() {
				policy = GenerateDynamicScaleOutPolicy(1, 2, "responsetime", 2000)
				initialInstanceCount = 1
			})

			JustBeforeEach(func() {
				ticker = time.NewTicker(10 * time.Second)
				go func(chan bool) {
					defer GinkgoRecover()
					for {
						select {
						case <-doneChan:
							ticker.Stop()
							doneAcceptChan <- true
							return
						case <-ticker.C:
							cfh.CurlApp(cfg, appName, "/responsetime/slow/3000", "-f")
						}
					}
				}(doneChan)
			})

			It("should scale out", Label(acceptance.LabelSmokeTests), func() {
				WaitForNInstancesRunning(appGUID, 2, 5*time.Minute)
			})
		})

		Context("when responsetime is less than scaling in threshold", func() {

			BeforeEach(func() {
				policy = GenerateDynamicScaleInPolicy(1, 2, "responsetime", 1000)
				initialInstanceCount = 2
			})

			JustBeforeEach(func() {
				ticker = time.NewTicker(2 * time.Second)
				go func(chan bool) {
					defer GinkgoRecover()
					for {
						select {
						case <-doneChan:
							ticker.Stop()
							doneAcceptChan <- true
							return
						case <-ticker.C:
							cfh.CurlApp(cfg, appName, "/responsetime/fast", "-f")
						}
					}
				}(doneChan)
			})

			It("should scale in", func() {
				WaitForNInstancesRunning(appGUID, 1, 5*time.Minute)
			})
		})

	})

	Context("when scaling by throughput", func() {

		JustBeforeEach(func() {
			doneChan = make(chan bool)
			doneAcceptChan = make(chan bool)
		})

		AfterEach(func() {
			close(doneChan)
			Eventually(doneAcceptChan, 10*time.Second).Should(Receive())
		})

		Context("when throughput is greater than scaling out threshold", func() {

			BeforeEach(func() {
				policy = GenerateDynamicScaleOutPolicy(1, 2, "throughput", 1)
				initialInstanceCount = 1
			})

			JustBeforeEach(func() {
				ticker = time.NewTicker(25 * time.Millisecond)
				go func(chan bool) {
					defer GinkgoRecover()
					for {
						select {
						case <-doneChan:
							ticker.Stop()
							doneAcceptChan <- true
							return
						case <-ticker.C:
							cfh.CurlApp(cfg, appName, "/responsetime/fast", "-f")
						}
					}
				}(doneChan)
			})

			It("should scale out", func() {
				WaitForNInstancesRunning(appGUID, 2, 5*time.Minute)
			})

		})

		Context("when throughput is less than scaling in threshold", func() {

			BeforeEach(func() {
				policy = GenerateDynamicScaleInPolicy(1, 2, "throughput", 100)
				initialInstanceCount = 2
			})

			JustBeforeEach(func() {
				ticker = time.NewTicker(10 * time.Second)
				go func(chan bool) {
					defer GinkgoRecover()
					for {
						select {
						case <-doneChan:
							ticker.Stop()
							doneAcceptChan <- true
							return
						case <-ticker.C:
							cfh.CurlApp(cfg, appName, "/responsetime/fast", "-f")
						}
					}
				}(doneChan)
			})
			It("should scale in", func() {
				WaitForNInstancesRunning(appGUID, 1, 5*time.Minute)
			})
		})

	})

	// To check existing aggregated cpu metrics do: cf asm APP_NAME cpu
	Context("when scaling by cpu", func() {

		BeforeEach(func() {
			policy = GenerateDynamicScaleOutAndInPolicy(1, 2, "cpu", int64(float64(cfg.CPUUpperThreshold)*0.2), int64(float64(cfg.CPUUpperThreshold)*0.4))
			initialInstanceCount = 1
		})

		It("when cpu is greater than scaling out threshold", func() {
			By("should scale out to 2 instances")
			AppSetCpuUsage(cfg, appName, int(float64(cfg.CPUUpperThreshold)*0.9), 5)
			WaitForNInstancesRunning(appGUID, 2, 5*time.Minute)

			By("should scale in to 1 instance after cpu usage is reduced")
			//only hit the one instance that was asked to run hot.
			AppEndCpuTest(cfg, appName, 0)

			WaitForNInstancesRunning(appGUID, 1, 10*time.Minute)
		})
	})

	Context("when there is a scaling policy for cpuutil", func() {
		BeforeEach(func() {
			policy = GenerateDynamicScaleOutAndInPolicy(1, 2, "cpuutil", 40, 80)
			initialInstanceCount = 1
		})

		It("should scale out and in", func() {
			// this test depends on the size of the diego cells, currently 8 CPU 32 GB RAM,
			// and the CPU entitlements per share configured in ci/operations/set-cpu-entitlement-per-share.yaml.
			// if any of these dependencies change, the test may need some adjustments as well.
			//
			// the following shows the calculations this test is based on:
			//   - diego cell = 8 CPU 32GB RAM
			//   - total shares = 1024 * 32[gb host ram] / 8[upper limit of app memory] = 4096
			//   - CPU entitlement per share = 8[number host CPUs] * 100/ 4096[total shares] = 0,1953
			//   - CPU entitlement = 4096[total shares] / (32[gb host ram] * 1024) * (1[app memory in GB] * 1024) * 0,1953 ~= 25%
			//
			// in a nutshell: 1GB app memory results in a maximum cpu entitlement of 25%,
			// this means that cpuutil will be 100% if app cpu is at 25%.

			SetAppMemory(cfg, appName, "1GB")
			cpuEntitlementOfAppWith1GBMemory := 25
			maxCPUUsage := cpuEntitlementOfAppWith1GBMemory

			AppSetCpuUsage(cfg, appName, maxCPUUsage, 5)
			WaitForNInstancesRunning(appGUID, 2, 5*time.Minute)

			//only hit the one instance that was asked to run hot.
			AppEndCpuTest(cfg, appName, 0)
			WaitForNInstancesRunning(appGUID, 1, 5*time.Minute)
		})
	})
})

func min(a int, b int) int {
	if a <= b {
		return a
	}
	return b
}
