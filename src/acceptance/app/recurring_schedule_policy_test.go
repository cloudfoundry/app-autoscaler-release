package app_test

import (
	"acceptance"
	. "acceptance/helpers"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AutoScaler recurring schedule policy", func() {
	var (
		initialInstanceCount int
		daysOfMonthOrWeek    Days
		startTime            time.Time
		endTime              time.Time
		policy               string
	)

	BeforeEach(func() {
		instanceName = CreateService(cfg)
		initialInstanceCount = 1
		appName = CreateTestApp(cfg, "recurring-schedule", initialInstanceCount)
		appGUID = GetAppGuid(cfg, appName)
	})
	AfterEach(AppAfterEach)

	Context("when scaling by recurring schedule", func() {

		JustBeforeEach(func() {
			startTime, endTime = getStartAndEndTime(time.UTC, 70*time.Second, time.Duration(interval+120)*time.Second)
			policy = GenerateDynamicAndRecurringSchedulePolicy(1, 4, 80, "UTC", startTime, endTime, daysOfMonthOrWeek, 2, 5, 3)
			instanceName = CreatePolicy(cfg, appName, appGUID, policy)
			StartApp(appName, cfg.CfPushTimeoutDuration())
		})

		scaleDown := func() {
			By("setting to initial_min_instance_count")
			jobRunTime := time.Until(startTime.Add(5 * time.Minute))
			WaitForNInstancesRunning(appGUID, 3, jobRunTime)

			By("setting schedule's instance_min_count")
			jobRunTime = time.Until(endTime)
			Eventually(func() int { return RunningInstances(appGUID, jobRunTime) }, jobRunTime, 15*time.Second).Should(Equal(2))

			By("setting to default instance_min_count")
			WaitForNInstancesRunning(appGUID, 1, time.Until(endTime.Add(time.Duration(interval+60)*time.Second)))
		}

		Context("with days of month", func() {
			BeforeEach(func() { daysOfMonthOrWeek = DaysOfMonth })
			for i := 0; i < 20; i++ {
				It(fmt.Sprintf("should scale %d", i), scaleDown)
			}
		})

		Context("with days of week", func() {
			BeforeEach(func() { daysOfMonthOrWeek = DaysOfWeek })
			for i := 0; i < 20; i++ {
				It(fmt.Sprintf("should scale %d", i), Label(acceptance.LabelSmokeTests), scaleDown)
			}
		})
	})
})
