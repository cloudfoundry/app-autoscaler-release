package app_test

import (
	"net/http"
	"sync"
	"time"

	"code.cloudfoundry.org/app-autoscaler-release/src/acceptance/assets/app/go_app/internal/app"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CPU tests", func() {

	Context("CPU handler", func() {
		var cpuUsed uint64
		var cpuRuntime time.Duration
		cpuLock := &sync.Mutex{}
		cpuLock.Lock()
		useCPUFn := func(utilization uint64, duration time.Duration) {
			cpuUsed = utilization
			cpuRuntime = duration
			cpuLock.Unlock()
		}
		It("should err if utilization not an int64", func() {
			apiTest(NoOpSleep, NoOpUseMem, NoOpUseCPU, NoOpPostCustomMetrics).
				Get("/cpu/invalid/4").
				Expect(GinkgoT()).
				Status(http.StatusBadRequest).
				Body(`{"error":{"description":"invalid utilization: strconv.ParseUint: parsing \"invalid\": invalid syntax"}}`).
				End()
		})
		It("should err if cpu out of bounds", func() {
			apiTest(NoOpSleep, NoOpUseMem, NoOpUseCPU, NoOpPostCustomMetrics).
				Get("/cpu/100001010101010249032897287298719874687936483275648273632429479827398798271/4").
				Expect(GinkgoT()).
				Status(http.StatusBadRequest).
				Body(`{"error":{"description":"invalid utilization: strconv.ParseUint: parsing \"100001010101010249032897287298719874687936483275648273632429479827398798271\": value out of range"}}`).
				End()
		})
		It("should err if cpu not an int", func() {
			apiTest(NoOpSleep, NoOpUseMem, NoOpUseCPU, NoOpPostCustomMetrics).
				Get("/cpu/5/invalid").
				Expect(GinkgoT()).
				Status(http.StatusBadRequest).
				Body(`{"error":{"description":"invalid minutes: strconv.ParseUint: parsing \"invalid\": invalid syntax"}}`).
				End()
		})
		It("should return ok and sleep correctDuration", func() {
			apiTest(NoOpSleep, NoOpUseMem, useCPUFn, NoOpPostCustomMetrics).
				Get("/cpu/5/4").
				Expect(GinkgoT()).
				Status(http.StatusOK).
				Body(`{"utilization":5, "minutes":4 }`).
				End()
			cpuLock.Lock()
			Expect(cpuRuntime).Should(Equal(4 * time.Minute))
			Expect(cpuUsed).Should(Equal(uint64(5)))
		})
	})
	Context("UseCPU", func() {
		It("should use cpu and release when stopped", func() {

			oldCpu := getTotalCPUUsage("before cpuTest info test")

			By("allocating cpu")
			cpuInfo := &app.CPUTest{}
			cpuInfo.UseCPU(100, time.Second)
			Expect(cpuInfo.IsRunning()).To(Equal(true))
			Eventually(cpuInfo.IsRunning, "2s").Should(Equal(false))
			newCpu := getTotalCPUUsage("after cpuTest info test")
			Expect(newCpu - oldCpu).To(BeNumerically(">=", 500*time.Millisecond))
		})
	})
})

func getTotalCPUUsage(action string) time.Duration {
	GinkgoHelper()

	proc := getProcessInfo()

	stat, err := proc.Stat()
	Expect(err).ToNot(HaveOccurred())

	result := time.Duration(stat.CPUTime() * float64(time.Second))
	GinkgoWriter.Printf("total cpu time %s: %s\n", action, result.String())

	return result
}
