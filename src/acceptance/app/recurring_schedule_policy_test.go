package app

import (
	"acceptance/config"
	"fmt"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

type days string

const(
	daysOfMonth days = "days_of_month"
	daysOfWeek       = "days_of_week"
)

var _ = Describe("AutoScaler recurring schedule policy", func() {
	var (
		appName              string
		appGUID              string
		instanceName         string
		initialInstanceCount int
		daysOfMonthOrWeek    days
		location             *time.Location
		startTime            time.Time
		endTime              time.Time
	)

	BeforeEach(func() {
		instanceName = generator.PrefixedRandomName("autoscaler", "service")
		createService := cf.Cf("create-service", cfg.ServiceName, cfg.ServicePlan, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(createService).To(Exit(0), "failed creating service")

		initialInstanceCount = 1
		appName = generator.PrefixedRandomName("autoscaler", "nodeapp")
		countStr := strconv.Itoa(initialInstanceCount)
		createApp := cf.Cf("push", appName, "--no-start", "-i", countStr, "-b", cfg.NodejsBuildpackName, "-m", cfg.NodeMemoryLimit, "-p", config.NODE_APP, "-d", cfg.AppsDomain).Wait(cfg.DefaultTimeoutDuration())
		Expect(createApp).To(Exit(0), "failed creating app")

		guid := cf.Cf("app", appName, "--guid").Wait(cfg.DefaultTimeout)
		Expect(guid).To(Exit(0))
		appGUID = strings.TrimSpace(string(guid.Out.Contents()))
	})

	AfterEach(func() {
		Expect(cf.Cf("delete", appName, "-f", "-r").Wait(cfg.CfPushTimeoutDuration())).To(Exit(0))

		deleteService := cf.Cf("delete-service", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
		Expect(deleteService).To(Exit(0))
	})

	Context("when scale out by recurring schedule", func() {

		JustBeforeEach(func() {
			timeZone := "GMT"
			location, _ = time.LoadLocation(timeZone)
			startTime, endTime = getStartAndEndTime(location)

			policyByte, err := ioutil.ReadFile("../assets/file/policy/recurringschedule.json")
			Expect(err).NotTo(HaveOccurred())

			policyStr := setRecurringScheduleDateTime(policyByte, timeZone, startTime, endTime, daysOfMonthOrWeek)
			bindService := cf.Cf("bind-service", appName, instanceName, "-c", policyStr).Wait(cfg.DefaultTimeoutDuration())
			Expect(bindService).To(Exit(0), "failed binding service to app with a policy ")

			Expect(cf.Cf("start", appName).Wait(cfg.DefaultTimeout * 2)).To(Exit(0))
			waitForNInstancesRunning(appGUID, initialInstanceCount, cfg.DefaultTimeoutDuration())
		})

		AfterEach(func() {
			unbindService := cf.Cf("unbind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
			Expect(unbindService).To(Exit(0), "failed unbinding service from app")
		})

		Context("with days of month", func() {
			BeforeEach(func() {
				daysOfMonthOrWeek = daysOfMonth
			})

			It("should scale", func() {
				totalTime := time.Duration(cfg.ReportInterval*2)*time.Second + 2*time.Minute

				By("setting to initial_min_instance_count")
				waitForNInstancesRunning(appGUID, 3, totalTime)

				By("setting schedule's instance_min_count")
				jobRunTime := endTime.Sub(time.Now().In(location))
				Eventually(func() int {
					return runningInstances(appGUID, jobRunTime)
				}, jobRunTime, 15*time.Second).Should(Equal(2))

				jobRunTime = endTime.Sub(time.Now().In(location))
				Consistently(func() int {
					return runningInstances(appGUID, jobRunTime)
				}, jobRunTime, 15*time.Second).Should(Equal(2))

				By("setting to default instance_min_count")
				waitForNInstancesRunning(appGUID, 1, totalTime)
			})

		})

		Context("with days of week", func() {
			BeforeEach(func() {
				daysOfMonthOrWeek = daysOfWeek
			})

			It("should scale", func() {
				totalTime := time.Duration(cfg.ReportInterval*2)*time.Second + 2*time.Minute

				By("setting to initial_min_instance_count")
				waitForNInstancesRunning(appGUID, 3, totalTime)

				By("setting schedule's instance_min_count")
				jobRunTime := endTime.Sub(time.Now().In(location))
				Eventually(func() int {
					return runningInstances(appGUID, jobRunTime)
				}, jobRunTime, 15*time.Second).Should(Equal(2))

				jobRunTime = endTime.Sub(time.Now().In(location))
				Consistently(func() int {
					return runningInstances(appGUID, jobRunTime)
				}, jobRunTime, 15*time.Second).Should(Equal(2))

				By("setting to default instance_min_count")
				waitForNInstancesRunning(appGUID, 1, totalTime)
			})
		})
	})

})

func getStartAndEndTime(location *time.Location) (time.Time, time.Time) {
	// Since the validation of time could fail if spread over two days and will result in acceptance test failure
	// Need to fix dates in that case.
	jobDuration := 4 * time.Minute;
	offset := 70 * time.Second
	startTime := time.Now().In(location).Add(offset).Truncate(time.Minute)


	if startTime.Day() != startTime.Add(jobDuration).Day() {
		startTime = startTime.Add(jobDuration).Truncate(24*time.Hour)
	}

	endTime := startTime.Add(jobDuration)
	return startTime, endTime
}

func setRecurringScheduleDateTime(policyByte []byte, timeZone string, startTime time.Time, endTime time.Time, daysOfMonthOrWeek days) string {
	var day int
	timeParseFormat := "15:04"
	startTimeStr := startTime.Format(timeParseFormat)
	endTimeStr := endTime.Format(timeParseFormat)
	if daysOfMonthOrWeek == daysOfMonth {
		day = startTime.Day()
	} else {
		day = int(startTime.Weekday())
		// 0 here is Sunday, scheduler expects 7 for Sunday
		if day == 0 {
			day = 7
		}
	}
	return fmt.Sprintf(string(policyByte), timeZone, startTimeStr, endTimeStr, daysOfMonthOrWeek, day)
}
