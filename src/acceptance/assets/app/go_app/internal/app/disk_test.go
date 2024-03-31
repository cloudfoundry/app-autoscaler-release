package app_test

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/app-autoscaler-release/src/acceptance/assets/app/go_app/internal/app"
	"code.cloudfoundry.org/app-autoscaler-release/src/acceptance/assets/app/go_app/internal/app/appfakes"
	"github.com/fgrosse/zaptest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/steinfletcher/apitest"
)

var _ = Describe("Disk handler", func() {

	var mockDiskOccupier *appfakes.FakeDiskOccupier

	apiTest := func(diskOccupier app.DiskOccupier) *apitest.APITest {
		GinkgoHelper()
		logger := zaptest.LoggerWriter(GinkgoWriter)

		return apitest.New().Handler(app.Router(logger, nil, nil, nil, diskOccupier, nil))
	}

	BeforeEach(func() {
		mockDiskOccupier = &appfakes.FakeDiskOccupier{}
	})

	It("should err if utilization not an int64", func() {
		apiTest(mockDiskOccupier).
			Get("/disk/invalid/4").
			Expect(GinkgoT()).
			Status(http.StatusBadRequest).
			Body(`{"error":{"description":"invalid utilization: strconv.ParseInt: parsing \"invalid\": invalid syntax"}}`).
			End()
	})
	It("should err if disk out of bounds", func() {
		apiTest(mockDiskOccupier).
			Get("/disk/100001010101010249032897287298719874687936483275648273632429479827398798271/4").
			Expect(GinkgoT()).
			Status(http.StatusBadRequest).
			Body(`{"error":{"description":"invalid utilization: strconv.ParseInt: parsing \"100001010101010249032897287298719874687936483275648273632429479827398798271\": value out of range"}}`).
			End()
	})
	It("should err if disk not an int", func() {
		apiTest(mockDiskOccupier).
			Get("/disk/5/invalid").
			Expect(GinkgoT()).
			Status(http.StatusBadRequest).
			Body(`{"error":{"description":"invalid minutes: strconv.ParseInt: parsing \"invalid\": invalid syntax"}}`).
			End()
	})
	It("should return ok", func() {
		apiTest(mockDiskOccupier).
			Get("/disk/100/2").
			Expect(GinkgoT()).
			Status(http.StatusOK).
			Body(`{"XXutilization":100, "minutes":2 }`).
			End()
	})
	It("should err if already running", func() {
		mockDiskOccupier.OccupyReturns(errors.New("already occupying"))
		apiTest(mockDiskOccupier).
			Get("/disk/100/2").
			Expect(GinkgoT()).
			Status(http.StatusInternalServerError).
			Body(`{"error":{"description":"error invoking occupation: already occupying"}}`).
			End()
	})
})

var _ = Describe("DefaultDiskOccupier", func() {

	var filePath string
	var oneHundredKB int64
	var duration time.Duration
	var do app.DiskOccupier

	BeforeEach(func() {
		filePath = filepath.Join(GinkgoT().TempDir(), "this-file-is-being-used-to-eat-up-the-disk")
		oneHundredKB = 100 * 1000 // 100 KB
		duration = 2 * time.Second
		do = app.NewDefaultDiskOccupier(filePath)
	})

	Describe("Occupy", func() {
		When("not occupying already", func() {
			It("occupies oneHundredKB for a certain amount of time", func() {
				err := do.Occupy(oneHundredKB, duration)
				Expect(err).ToNot(HaveOccurred())

				fStat, err := os.Stat(filePath)
				Expect(err).ToNot(HaveOccurred())
				Expect(fStat.Size()).To(Equal(oneHundredKB))

				Eventually(func() bool {
					return isGone(filePath)
				}, 2*duration, 50*time.Millisecond)
			})
		})

		When("occupying already started", func() {
			It("fails with an error", func() {
				// initial occupation
				err := do.Occupy(oneHundredKB, duration)
				Expect(err).ToNot(HaveOccurred())

				// try to occupy again
				err = do.Occupy(oneHundredKB, duration)
				Expect(err).To(MatchError(errors.New("disk space is already being occupied")))
			})
		})

		When("occupation just ended", func() {
			It("is possible to start occupy again", func() {
				// initial occupation
				veryShortTime := 10 * time.Millisecond
				err := do.Occupy(oneHundredKB, veryShortTime)
				Expect(err).ToNot(HaveOccurred())

				// wait till occupation is over
				time.Sleep(2 * veryShortTime)

				// occupy again
				err = do.Occupy(oneHundredKB, veryShortTime)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Stop", func() {
		When("it is occupying already", func() {
			It("stops occupying oneHundredKB", func() {
				tremendousAmountOfTime := 999999999 * duration
				err := do.Occupy(oneHundredKB, tremendousAmountOfTime)
				Expect(err).ToNot(HaveOccurred())

				do.Stop()

				Expect(isGone(filePath))
			})
		})

		When("it is not occupying already", func() {
			It("does nothing", func() {
				do.Stop()

				Expect(true)
			})
		})

		When("occupation just ended", func() {
			It("does nothing", func() {
				veryShortTime := 10 * time.Millisecond
				err := do.Occupy(oneHundredKB, veryShortTime)
				Expect(err).ToNot(HaveOccurred())

				// wait till occupation is over
				time.Sleep(2 * veryShortTime)

				do.Stop()

				Expect(true)
			})
		})
	})
})

func isGone(filePath string) bool {
	gone := false
	if _, err := os.Stat(filePath); err != nil && errors.Is(err, os.ErrNotExist) {
		gone = true
	}
	return gone
}
