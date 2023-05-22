package app_test

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"code.cloudfoundry.org/app-autoscaler-release/src/acceptance/assets/app/go_app/internal/app"
	"code.cloudfoundry.org/app-autoscaler-release/src/acceptance/assets/app/go_app/internal/app/appfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/procfs"
)

var _ = Describe("Memory tests", func() {

	Context("Memory tests", func() {

		fakeMemoryTest := &appfakes.FakeMemoryGobbler{}

		It("should err if memory not an int64", func() {
			apiTest(nil, fakeMemoryTest, nil, nil).
				Get("/memory/invalid/4").
				Expect(GinkgoT()).
				Status(http.StatusBadRequest).
				Body(`{"error":{"description":"invalid memoryMiB: strconv.ParseUint: parsing \"invalid\": invalid syntax"}}`).
				End()
		})
		It("should err if memory out of bounds", func() {
			apiTest(nil, fakeMemoryTest, nil, nil).
				Get("/memory/100001010101010249032897287298719874687936483275648273632429479827398798271/4").
				Expect(GinkgoT()).
				Status(http.StatusBadRequest).
				Body(`{"error":{"description":"invalid memoryMiB: strconv.ParseUint: parsing \"100001010101010249032897287298719874687936483275648273632429479827398798271\": value out of range"}}`).
				End()
		})
		It("should err if memory not an int", func() {
			apiTest(nil, fakeMemoryTest, nil, nil).
				Get("/memory/5/invalid").
				Expect(GinkgoT()).
				Status(http.StatusBadRequest).
				Body(`{"error":{"description":"invalid minutes: strconv.ParseUint: parsing \"invalid\": invalid syntax"}}`).
				End()
		})
		It("should return ok and sleep correctDuration", func() {
			apiTest(nil, fakeMemoryTest, nil, nil).
				Get("/memory/5/4").
				Expect(GinkgoT()).
				Status(http.StatusOK).
				Body(`{"memoryMiB":5, "minutes":4 }`).
				End()

			// In the following lines we sometimes use “Eventually” instead of “Expect” to ensure that
			// these checks are run after the asynchronous go-routines in memory.go.MemoryTests
			// have been finished. Usually this would require making use of asynchronous
			// http-features (i.e. returing http-status-code 202 etc.) which would involve a lot of
			// new lines of code.
			Eventually(func() int { return fakeMemoryTest.UseMemoryCallCount() }).Should(Equal(1))
			Expect(fakeMemoryTest.UseMemoryArgsForCall(0)).To(Equal(uint64(5 * app.Mebi)))
			Eventually(func() int { return fakeMemoryTest.SleepCallCount() }).Should(Equal(1))
			Expect(fakeMemoryTest.SleepArgsForCall(0)).To(Equal(4 * time.Minute))
		})
	})
	Context("memTest info tests", func() {
		It("should gobble memory and release when stopped", func() {
			var allocInMebi uint64 = 50 * app.Mebi

			oldMem := getTotalMemoryUsage("before memTest info test")
			slack := getMemorySlack()

			By("allocating memory")
			memInfo := &app.ListBasedMemoryGobbler{}
			memInfo.UseMemory(allocInMebi)
			Expect(memInfo.IsRunning()).To(Equal(true))

			newMem := getTotalMemoryUsage("during memTest info test")
			msg :=
			`
			If this test fails, please consider to rewrite internal/app/memory.go.UseMemory()
			to make use of syscalls directly (e.g. “malloc”) to make sure of not
			having issues due to the go-runtime.
			`
			GinkgoWriter.Printf(msg)
			Expect(newMem - oldMem).To(BeNumerically(">=", allocInMebi - slack))

			By("and releasing it after the test ends")
			memInfo.StopTest()
			Expect(memInfo.IsRunning()).To(Equal(false))

			slack = getMemorySlack()

			Eventually(getTotalMemoryUsage).WithArguments("after memTest info test").Should(BeNumerically("<=", newMem - allocInMebi + slack))
		})
	})
})

func getTotalMemoryUsage(action string) uint64 {
	GinkgoHelper()

	runtime.GC()
	proc := getProcessInfo()

	stat, err := proc.NewStatus()
	Expect(err).ToNot(HaveOccurred())

	result := stat.VmRSS + stat.VmSwap

	GinkgoWriter.Printf("total memory usage %s: %d MiB\n", action, result / app.Mebi)

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

	slack := ms.HeapInuse - ms.HeapAlloc

	GinkgoWriter.Printf("slack: %d MiB\n", slack / app.Mebi)

	return slack
}

func getProcessInfo() procfs.Proc {
	GinkgoHelper()
	fs, err := procfs.NewFS("/proc")
	Expect(err).ToNot(HaveOccurred())

	proc, err := fs.Proc(os.Getpid())
	Expect(err).ToNot(HaveOccurred())

	return proc
}
