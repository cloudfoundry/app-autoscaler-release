package app_test

import (
	"acceptance/assets/app/go_app/internal/app"
	"net/http"
	"runtime"
	"runtime/debug"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Memory tests", func() {

	Context("Memory tests", func() {
		var amountSlept time.Duration
		var memUsed uint64
		sleepFn := func(duration time.Duration) { amountSlept = duration }
		useMemFn := func(useMb uint64) { memUsed = useMb }
		It("should err if memory not an int64", func() {
			apiTest(sleepFn, useMemFn).
				Get("/memory/invalid/4").
				Expect(GinkgoT()).
				Status(http.StatusBadRequest).
				Body(`{"error":{"description":"invalid memoryMb: strconv.ParseUint: parsing \"invalid\": invalid syntax"}}`).
				End()
		})
		It("should err if memory out of bounds", func() {
			apiTest(sleepFn, useMemFn).
				Get("/memory/100001010101010249032897287298719874687936483275648273632429479827398798271/4").
				Expect(GinkgoT()).
				Status(http.StatusBadRequest).
				Body(`{"error":{"description":"invalid memoryMb: strconv.ParseUint: parsing \"100001010101010249032897287298719874687936483275648273632429479827398798271\": value out of range"}}`).
				End()
		})
		It("should err if memory not an int", func() {
			apiTest(sleepFn, useMemFn).
				Get("/memory/5/invalid").
				Expect(GinkgoT()).
				Status(http.StatusBadRequest).
				Body(`{"error":{"description":"invalid minutes: strconv.ParseUint: parsing \"invalid\": invalid syntax"}}`).
				End()
		})
		It("should return ok and sleep correctDuration", func() {
			apiTest(sleepFn, useMemFn).
				Get("/memory/5/4").
				Expect(GinkgoT()).
				Status(http.StatusOK).
				Body(`{"memoryMb":5, "minutes":4 }`).
				End()
			Eventually(amountSlept).Should(Equal(4 * time.Minute))
			Eventually(memUsed).Should(Equal(uint64(5)))
		})
	})
	Context("memTest info tests", func() {
		It("should gobble memory and release when stopped", func() {
			memInfo := &app.MemTestInfo{}
			runtime.GC()
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)
			memInfo.UseMemory(5 * app.Megabyte)
			runtime.GC()

			var msNew runtime.MemStats
			runtime.ReadMemStats(&msNew)
			Expect(memInfo.IsRunning()).To(Equal(true))
			Expect(msNew.HeapInuse - ms.HeapInuse).To(BeNumerically(">=", 5*app.Megabyte))

			memInfo.StopTest()
			runtime.GC()
			runtime.ReadMemStats(&ms)
			Expect(memInfo.IsRunning()).To(Equal(false))
			Eventually(func() uint64 {
				var stat runtime.MemStats
				runtime.GC()
				debug.FreeOSMemory()
				runtime.ReadMemStats(&stat)
				return msNew.HeapInuse - stat.HeapInuse
			}).Should(BeNumerically(">=", 5*app.Megabyte))

		})
	})
})
