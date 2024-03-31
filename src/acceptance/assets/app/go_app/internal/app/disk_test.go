package app_test

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/app-autoscaler-release/src/acceptance/assets/app/go_app/internal/app"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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
		When("it is not occupying already", func() {
			FIt("occupies oneHundredKB for a certain amount of time", func() {
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

		When("it is occupying already", func() {
			FIt("fails with an error", func() {
				// initial occupation
				err := do.Occupy(oneHundredKB, duration)
				Expect(err).ToNot(HaveOccurred())

				// try to occupy again
				err = do.Occupy(oneHundredKB, duration)
				Expect(err).To(MatchError(errors.New("disk space is already being occupied")))
			})
		})

		When("an occupation just ended", func() {
			FIt("is possible to start occupy again", func() {
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
			FIt("stops occupying oneHundredKB", func() {
				tremendousAmountOfTime := 999999999 * duration
				err := do.Occupy(oneHundredKB, tremendousAmountOfTime)
				Expect(err).ToNot(HaveOccurred())

				do.Stop()

				Expect(isGone(filePath))
			})
		})

		When("it is not occupying already", func() {
			FIt("does nothing", func() {
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
