package app_test

import (
	"net/http"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Responsetime tests", func() {

	Context("Responsetime tests", func() {
		var amountSlept time.Duration
		sleepLock := &sync.Mutex{}
		sleepLock.Lock()
		sleepFn := func(duration time.Duration) { amountSlept = duration; sleepLock.Unlock() }
		It("should err if delayInMS not an int64", func() {
			apiTest(sleepFn, NoOpUseMem, NoOpUseCPU, NoOpPostCustomMetrics).
				Get("/responsetime/slow/yes").
				Expect(GinkgoT()).
				Status(http.StatusBadRequest).
				Body(`{"error":{"description":"invalid milliseconds: strconv.ParseUint: parsing \"yes\": invalid syntax"}}`).
				End()
		})
		It("should err if memory out of bounds", func() {
			apiTest(sleepFn, NoOpUseMem, NoOpUseCPU, NoOpPostCustomMetrics).
				Get("/responsetime/slow/100001010101010249032897287298719874687936483275648273632429479827398798271").
				Expect(GinkgoT()).
				Status(http.StatusBadRequest).
				Body(`{"error":{"description":"invalid milliseconds: strconv.ParseUint: parsing \"100001010101010249032897287298719874687936483275648273632429479827398798271\": value out of range"}}`).
				End()
		})

		It("should return ok and sleep correctDuration", func() {
			apiTest(sleepFn, NoOpUseMem, NoOpUseCPU, NoOpPostCustomMetrics).
				Get("/responsetime/slow/4000").
				Expect(GinkgoT()).
				Status(http.StatusOK).
				Body(`{"duration":"4s"}`).
				End()
			sleepLock.Lock()
			Expect(amountSlept).Should(Equal(4000 * time.Millisecond))
		})
	})
})
