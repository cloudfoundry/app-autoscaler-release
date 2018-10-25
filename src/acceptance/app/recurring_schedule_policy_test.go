package app

import (
	"acceptance/config"
	. "acceptance/helpers"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"strconv"
	"strings"
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
		if cfg.IsServiceOfferingEnabled() {
			instanceName = generator.PrefixedRandomName("autoscaler", "service")
			createService := cf.Cf("create-service", cfg.ServiceName, cfg.ServicePlan, instanceName).Wait(cfg.DefaultTimeoutDuration())
			Expect(createService).To(Exit(0), "failed creating service")
		}

		initialInstanceCount = 1
		appName = generator.PrefixedRandomName("autoscaler", "nodeapp")
		countStr := strconv.Itoa(initialInstanceCount)
		createApp := cf.Cf("push", appName, "--no-start", "-i", countStr, "-b", cfg.NodejsBuildpackName, "-m", "128M", "-p", config.NODE_APP, "-d", cfg.AppsDomain).Wait(cfg.CfPushTimeoutDuration())
		Expect(createApp).To(Exit(0), "failed creating app")

		guid := cf.Cf("app", appName, "--guid").Wait(cfg.DefaultTimeoutDuration())
		Expect(guid).To(Exit(0))
		appGUID = strings.TrimSpace(string(guid.Out.Contents()))
	})

	AfterEach(func() {
		DeletePolicy(appName, appGUID)
		Expect(cf.Cf("delete", appName, "-f", "-r").Wait(cfg.DefaultTimeoutDuration())).To(Exit(0))
	})

	Context("when scaling by recurring schedule", func() {

		JustBeforeEach(func() {

			Expect(cf.Cf("start", appName).Wait(cfg.CfPushTimeoutDuration())).To(Exit(0))
			WaitForNInstancesRunning(appGUID, initialInstanceCount, cfg.DefaultTimeoutDuration())

			location, err := time.LoadLocation("GMT")
			Expect(err).NotTo(HaveOccurred())
			startTime, endTime = getStartAndEndTime(location, 70*time.Second, time.Duration(interval+120)*time.Second)
			policy = GenerateDynamicAndRecurringSchedulePolicy(cfg, 1, 4, 80, "GMT", startTime, endTime, daysOfMonthOrWeek, 2, 5, 3)

			CreatePolicy(appName, appGUID, policy)
		})

		Context("with days of month", func() {
			BeforeEach(func() {
				daysOfMonthOrWeek = DaysOfMonth
			})

			It("should scale", func() {
				By("setting to initial_min_instance_count")
				jobRunTime := startTime.Add(1 * time.Minute).Sub(time.Now())
				WaitForNInstancesRunning(appGUID, 3, jobRunTime)

				By("setting schedule's instance_min_count")
				jobRunTime = endTime.Sub(time.Now())
				Eventually(func() int {
					return RunningInstances(appGUID, jobRunTime)
				}, jobRunTime, 15*time.Second).Should(Equal(2))

				jobRunTime = endTime.Sub(time.Now())
				Consistently(func() int {
					return RunningInstances(appGUID, jobRunTime)
				}, jobRunTime, 15*time.Second).Should(Equal(2))

				By("setting to default instance_min_count")
				WaitForNInstancesRunning(appGUID, 1, time.Duration(interval+60)*time.Second)
			})

		})

		Context("with days of week", func() {
			BeforeEach(func() {
				daysOfMonthOrWeek = DaysOfWeek
			})

			It("should scale", func() {
				By("setting to initial_min_instance_count")
				jobRunTime := startTime.Add(1 * time.Minute).Sub(time.Now())
				WaitForNInstancesRunning(appGUID, 3, jobRunTime)

				By("setting schedule's instance_min_count")
				jobRunTime = endTime.Sub(time.Now())
				Eventually(func() int {
					return RunningInstances(appGUID, jobRunTime)
				}, jobRunTime, 15*time.Second).Should(Equal(2))

				jobRunTime = endTime.Sub(time.Now())
				Consistently(func() int {
					return RunningInstances(appGUID, jobRunTime)
				}, jobRunTime, 15*time.Second).Should(Equal(2))

				By("setting to default instance_min_count")
				WaitForNInstancesRunning(appGUID, 1, time.Duration(interval+120)*time.Second)
			})
		})
	})

})
