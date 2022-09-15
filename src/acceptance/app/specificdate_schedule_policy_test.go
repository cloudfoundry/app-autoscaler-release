package app_test

import (
	. "acceptance/helpers"
	"fmt"
	"os"

	"time"

	"github.com/KevinJCross/cf-test-helpers/v2/cf"
	"github.com/KevinJCross/cf-test-helpers/v2/generator"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("AutoScaler specific date schedule policy", func() {
	var (
		appName              string
		appGUID              string
		initialInstanceCount int
		startDateTime        time.Time
		endDateTime          time.Time
		policy               string
	)

	BeforeEach(func() {

		if cfg.IsServiceOfferingEnabled() {
			instanceName = generator.PrefixedRandomName(cfg.Prefix, cfg.InstancePrefix)
			createService := cf.Cf("create-service", cfg.ServiceName, cfg.ServicePlan, instanceName, "-b", cfg.ServiceBroker).Wait(cfg.DefaultTimeoutDuration())
			Expect(createService).To(Exit(0), "failed creating service")
		}

		initialInstanceCount = 1
		appName = CreateTestApp(cfg, "date-schedule", initialInstanceCount)
		appGUID = GetAppGuid(cfg, appName)
	})

	AfterEach(func() {
		if os.Getenv("SKIP_TEARDOWN") == "true" {
			fmt.Println("Skipping Teardown...")
		} else {
			DeletePolicy(appName, appGUID)
			Expect(cf.Cf("delete", appName, "-f", "-r").Wait(cfg.DefaultTimeoutDuration())).To(Exit(0))
		}
	})

	Context("when scaling by specific date schedule", func() {
		const scheduleInstanceMin = 2
		const scheduleInstanceMax = 5
		const scheduledInstanceInit = 3
		JustBeforeEach(func() {
			//TODO the start app needs to be after the binding but the timings require the app been up already.
			StartApp(appName, cfg.CfPushTimeoutDuration())
			startDateTime = time.Now().In(time.UTC).Add(1 * time.Minute)
			endDateTime = startDateTime.Add(time.Duration(interval+120) * time.Second)

			policy = GenerateDynamicAndSpecificDateSchedulePolicy(1, 4, 80, "UTC", startDateTime, endDateTime, scheduleInstanceMin, scheduleInstanceMax, scheduledInstanceInit)
			instanceName = CreatePolicy(cfg, appName, appGUID, policy)
		})

		It("should scale", func() {
			By(fmt.Sprintf("waiting for scheduledInstanceInit: %d", scheduledInstanceInit))
			jobRunTime := time.Until(startDateTime.Add(1 * time.Minute))
			WaitForNInstancesRunning(appGUID, 3, jobRunTime)

			By(fmt.Sprintf("waiting for scheduleInstanceMin: %d", scheduleInstanceMin))
			jobRunTime = time.Until(endDateTime)
			Eventually(func() int { return RunningInstances(appGUID, jobRunTime) }).
				WithTimeout(jobRunTime).
				WithPolling(15 * time.Second).
				Should(Equal(2))

			jobRunTime = time.Until(endDateTime)
			By(fmt.Sprintf("waiting till end of schedule %dS and should stay %d instances", int(jobRunTime.Seconds()), scheduleInstanceMin))
			Consistently(func() int { return RunningInstances(appGUID, jobRunTime) }).
				WithTimeout(jobRunTime).
				WithPolling(15 * time.Second).
				Should(Equal(2))

			WaitForNInstancesRunning(appGUID, 1, time.Duration(interval+60)*time.Second)

		})
	})

})
