package app_test

import (
	"acceptance"
	. "acceptance/helpers"
	"os"

	"github.com/KevinJCross/cf-test-helpers/v2/cf"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"time"
)

var _ = Describe("AutoScaler recurring schedule policy", func() {
	var (
		appName              string
		appGUID              string
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

	AfterEach(func() {
		if os.Getenv("SKIP_TEARDOWN") == "true" {
			GinkgoWriter.Println("Skipping Teardown...")
		} else {
			DeletePolicy(appName, appGUID)
			Expect(cf.Cf("delete", appName, "-f", "-r").Wait(cfg.DefaultTimeoutDuration())).To(Exit(0))
		}
	})

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
			It("should scale", scaleDown)
		})

		Context("with days of week", func() {
			BeforeEach(func() { daysOfMonthOrWeek = DaysOfWeek })
			It("should scale", Label(acceptance.LabelSmokeTests), scaleDown)
		})
	})

})
