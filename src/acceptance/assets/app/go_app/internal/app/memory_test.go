package app_test

import (
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"code.cloudfoundry.org/app-autoscaler-release/src/acceptance/assets/app/go_app/internal/app"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/procfs"
)

var _ = Describe("Memory tests", func() {

	Context("Memory tests", func() {
		var amountSlept time.Duration
		var memUsed uint64
		memLock := &sync.Mutex{}
		memLock.Lock()
		sleepLock := &sync.Mutex{}
		sleepLock.Lock()
		sleepFn := func(duration time.Duration) { amountSlept = duration; sleepLock.Unlock() }
		useMemFn := func(useMb uint64) { memUsed = useMb; memLock.Unlock() }
		It("should err if memory not an int64", func() {
			apiTest(sleepFn, useMemFn, NoOpUseCPU, NoOpPostCustomMetrics).
				Get("/memory/invalid/4").
				Expect(GinkgoT()).
				Status(http.StatusBadRequest).
				Body(`{"error":{"description":"invalid memoryMiB: strconv.ParseUint: parsing \"invalid\": invalid syntax"}}`).
				End()
		})
		It("should err if memory out of bounds", func() {
			apiTest(sleepFn, useMemFn, NoOpUseCPU, NoOpPostCustomMetrics).
				Get("/memory/100001010101010249032897287298719874687936483275648273632429479827398798271/4").
				Expect(GinkgoT()).
				Status(http.StatusBadRequest).
				Body(`{"error":{"description":"invalid memoryMiB: strconv.ParseUint: parsing \"100001010101010249032897287298719874687936483275648273632429479827398798271\": value out of range"}}`).
				End()
		})
		It("should err if memory not an int", func() {
			apiTest(sleepFn, useMemFn, NoOpUseCPU, NoOpPostCustomMetrics).
				Get("/memory/5/invalid").
				Expect(GinkgoT()).
				Status(http.StatusBadRequest).
				Body(`{"error":{"description":"invalid minutes: strconv.ParseUint: parsing \"invalid\": invalid syntax"}}`).
				End()
		})
		It("should return ok and sleep correctDuration", func() {
			apiTest(sleepFn, useMemFn, NoOpUseCPU, NoOpPostCustomMetrics).
				Get("/memory/5/4").
				Expect(GinkgoT()).
				Status(http.StatusOK).
				Body(`{"memoryMiB":5, "minutes":4 }`).
				End()
			sleepLock.Lock()
			Expect(amountSlept).Should(Equal(4 * time.Minute))
			memLock.Lock()
			Expect(memUsed).Should(Equal(uint64(5)))
		})
	})
	Context("memTest info tests", func() {
		It("should gobble memory and release when stopped", func() {

			oldMem := getTotalMemoryUsage("before memTest info test")
			slack := getMemorySlack()

			By("allocating memory")
			memInfo := &app.MemTest{}
			memInfo.UseMemory(5 * app.Mebi)
			Expect(memInfo.IsRunning()).To(Equal(true))

			newMem := getTotalMemoryUsage("during memTest info test")
			Expect(newMem - oldMem).To(BeNumerically(">=", 5*app.Mebi-slack))

			By("and releasing it after the test ends")
			memInfo.StopTest()
			Expect(memInfo.IsRunning()).To(Equal(false))

			slack = getMemorySlack()
			GinkgoWriter.Printf("slack: %d MiB\n", slack/app.Mebi)

			Eventually(getTotalMemoryUsage).WithArguments("after memTest info test").Should(BeNumerically("<=", newMem-5*app.Mebi+slack))
		})
	})
})

func getTotalMemoryUsage(action string) uint64 {
	GinkgoHelper()

	proc := getProcessInfo()

	stat, err := proc.NewStatus()
	Expect(err).ToNot(HaveOccurred())

	result := stat.VmRSS + stat.VmSwap

	GinkgoWriter.Printf("total memory usage %s: %d MiB\n", action, result/app.Mebi)

	return result
}

// HeapInuse minus HeapAlloc estimates the amount of memory
// that has been dedicated to particular size classes, but is
// not currently being used. This is an upper bound on
// fragmentation, but in general this memory can be reused
// efficiently.
func getMemorySlack() uint64 {
	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	return ms.HeapInuse - ms.HeapAlloc
}

func getProcessInfo() procfs.Proc {
	GinkgoHelper()
	fs, err := procfs.NewFS("/proc")
	Expect(err).ToNot(HaveOccurred())

	proc, err := fs.Proc(os.Getpid())
	Expect(err).ToNot(HaveOccurred())

	return proc
}
