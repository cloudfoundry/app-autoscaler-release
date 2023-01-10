package app_test

import (
	. "acceptance/helpers"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AutoScaler specific date schedule policy", func() {
	var (
		initialInstanceCount int
		startDateTime        time.Time
		endDateTime          time.Time
		policy               string
		err                  error
	)

	BeforeEach(func() {
		instanceName = CreateService(cfg)
		initialInstanceCount = 1
		appName = CreateTestApp(cfg, "date-schedule", initialInstanceCount)
		appGUID, err = GetAppGuid(cfg, appName)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(AppAfterEach)

	Context("when scaling by specific date schedule", func() {
		const scheduleInstanceMin = 2
		const scheduleInstanceMax = 5
		const scheduledInstanceInit = 3
		JustBeforeEach(func() {
			//TODO the start app needs to be after the binding but the timings require the app been up already.
			StartApp(appName, cfg.CfPushTimeoutDuration())
			startDateTime = time.Now().In(time.UTC).Add(1 * time.Minute)
			endDateTime = startDateTime.Add(time.Duration(interval+120) * time.Second)

			policy = GenerateSpecificDateSchedulePolicy(startDateTime, endDateTime, scheduleInstanceMin, scheduleInstanceMax, scheduledInstanceInit)
			instanceName = CreatePolicy(cfg, appName, appGUID, policy)
		})

		It("should scale", func() {
			pollTime := 15 * time.Second
			By(fmt.Sprintf("waiting for scheduledInstanceInit: %d", scheduledInstanceInit))
			WaitForNInstancesRunning(appGUID, 3, time.Until(startDateTime.Add(1*time.Minute)))

			By(fmt.Sprintf("waiting for scheduleInstanceMin: %d", scheduleInstanceMin))
			jobRunTime := time.Until(endDateTime)
			Eventually(func() (int, error) { return RunningInstances(appGUID, cfg.DefaultTimeoutDuration()) }).
				//+/- poll time error margin.
				WithTimeout(jobRunTime + pollTime).
				WithPolling(pollTime).
				Should(Equal(2))

			//+/- poll time error margin.
			timeout := time.Until(endDateTime) - pollTime
			By(fmt.Sprintf("waiting till end of schedule %dS and should stay %d instances", int(timeout.Seconds()), scheduleInstanceMin))
			Consistently(func() (int, error) { return RunningInstances(appGUID, jobRunTime) }).
				WithTimeout(timeout).
				WithPolling(pollTime).
				Should(Equal(2))

			WaitForNInstancesRunning(appGUID, 1, time.Duration(interval+60)*time.Second)
		})
	})
})
