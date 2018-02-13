package app

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	"acceptance/config"
	. "acceptance/helpers"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type Days string

const (
	daysOfMonth Days = "days_of_month"
	daysOfWeek       = "days_of_week"
)

type appSummary struct {
	RunningInstances int `json:"running_instances"`
}

type ScalingPolicy struct {
	InstanceMin  int               `json:"instance_min_count"`
	InstanceMax  int               `json:"instance_max_count"`
	ScalingRules []*ScalingRule    `json:"scaling_rules,omitempty"`
	Schedules    *ScalingSchedules `json:"schedules,omitempty"`
}

type ScalingRule struct {
	MetricType            string `json:"metric_type"`
	StatWindowSeconds     int    `json:"stat_window_secs"`
	BreachDurationSeconds int    `json:"breach_duration_secs"`
	Threshold             int64  `json:"threshold"`
	Operator              string `json:"operator"`
	CoolDownSeconds       int    `json:"cool_down_secs"`
	Adjustment            string `json:"adjustment"`
}

type ScalingSchedules struct {
	Timezone              string                  `json:"timezone,omitempty"`
	RecurringSchedules    []*RecurringSchedule    `json:"recurring_schedule,omitempty"`
	SpecificDateSchedules []*SpecificDateSchedule `json:"specific_date,omitempty"`
}

type RecurringSchedule struct {
	StartTime             string `json:"start_time"`
	EndTime               string `json:"end_time"`
	DaysOfWeek            []int  `json:"days_of_week,omitempty"`
	DaysOfMonth           []int  `json:"days_of_month,omitempty"`
	ScheduledInstanceMin  int    `json:"instance_min_count"`
	ScheduledInstanceMax  int    `json:"instance_max_count"`
	ScheduledInstanceInit int    `json:"initial_min_instance_count"`
}

type SpecificDateSchedule struct {
	StartDateTime         string `json:"start_date_time"`
	EndDateTime           string `json:"end_date_time"`
	ScheduledInstanceMin  int    `json:"instance_min_count"`
	ScheduledInstanceMax  int    `json:"instance_max_count"`
	ScheduledInstanceInit int    `json:"initial_min_instance_count"`
}

const MB = 1024 * 1024

var (
	cfg      *config.Config
	setup    *workflowhelpers.ReproducibleTestSuiteSetup
	interval int
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)

	cfg = config.LoadConfig(t)
	componentName := "Application Scale Suite"
	rs := []Reporter{}

	if cfg.GetArtifactsDirectory() != "" {
		helpers.EnableCFTrace(cfg, componentName)
		rs = append(rs, helpers.NewJUnitReporter(cfg, componentName))
	}

	RunSpecsWithDefaultAndCustomReporters(t, componentName, rs)
}

var _ = BeforeSuite(func() {

	setup = workflowhelpers.NewTestSuiteSetup(cfg)
	setup.Setup()

	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		EnableServiceAccess(cfg, setup.GetOrganizationName())
	})

	serviceExists := cf.Cf("marketplace", "-s", cfg.ServiceName).Wait(cfg.DefaultTimeoutDuration())
	Expect(serviceExists).To(Exit(0), fmt.Sprintf("Service offering, %s, does not exist", cfg.ServiceName))

	interval = cfg.AggregateInterval
	if interval < 60 {
		interval = 60
	}

})

var _ = AfterSuite(func() {
	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		DisableServiceAccess(cfg, setup.GetOrganizationName())
	})
	setup.Teardown()
})

func runningInstances(appGUID string, timeout time.Duration) int {
	cmd := cf.Cf("curl", "/v2/apps/"+appGUID+"/summary")
	Expect(cmd.Wait(timeout)).To(Exit(0))

	var summary appSummary
	err := json.Unmarshal(cmd.Out.Contents(), &summary)
	Expect(err).ToNot(HaveOccurred())
	return summary.RunningInstances
}

func waitForNInstancesRunning(appGUID string, instances int, timeout time.Duration) {
	Eventually(func() int {
		return runningInstances(appGUID, timeout)
	}, timeout, 10*time.Second).Should(Equal(instances))
}

type instanceStats struct {
	MemQuota uint64 `json:"mem_quota"`
	Usage    instanceUsage
}

type instanceUsage struct {
	Mem uint64
}

type instanceInfo struct {
	State string
	Stats instanceStats
}

type appStats map[string]*instanceInfo

func memoryUsed(appGUID string, index int, timeout time.Duration) (uint64, uint64) {
	cmd := cf.Cf("curl", "/v2/apps/"+appGUID+"/stats")
	Expect(cmd.Wait(timeout)).To(Exit(0))

	var stats appStats
	err := json.Unmarshal(cmd.Out.Contents(), &stats)
	Expect(err).ToNot(HaveOccurred())

	instance := stats[strconv.Itoa(index)]
	if instance == nil {
		return 0, 0
	}

	return instance.Stats.Usage.Mem, instance.Stats.MemQuota
}

func allInstancesMemoryUsed(appGUID string, timeout time.Duration) []uint64 {
	cmd := cf.Cf("curl", "/v2/apps/"+appGUID+"/stats")
	Expect(cmd.Wait(timeout)).To(Exit(0))

	var stats appStats
	err := json.Unmarshal(cmd.Out.Contents(), &stats)
	Expect(err).ToNot(HaveOccurred())

	if len(stats) == 0 {
		return []uint64{}
	}

	mem := make([]uint64, len(stats))

	for k, instance := range stats {
		i, err := strconv.Atoi(k)
		Expect(err).NotTo(HaveOccurred())
		mem[i] = instance.Stats.Usage.Mem
	}
	return mem
}

func averageMemoryUsedByInstance(appGUID string, timeout time.Duration) uint64 {
	memoryUsedArray := allInstancesMemoryUsed(appGUID, timeout)
	instanceCount := len(memoryUsedArray)
	if instanceCount == 0 {
		return math.MaxInt64
	}

	var memSum uint64
	for _, m := range memoryUsedArray {
		memSum += m
	}

	return memSum / uint64(len(memoryUsedArray))
}

func generateDynamicScaleOutPolicy(instanceMin, instanceMax int, metricName string, threshold int64) string {
	scalingOutRule := ScalingRule{
		MetricType:            metricName,
		StatWindowSeconds:     interval,
		BreachDurationSeconds: interval,
		Threshold:             threshold,
		Operator:              ">=",
		CoolDownSeconds:       interval,
		Adjustment:            "+1",
	}

	policy := ScalingPolicy{
		InstanceMin:  instanceMin,
		InstanceMax:  instanceMax,
		ScalingRules: []*ScalingRule{&scalingOutRule},
	}
	bytes, err := json.Marshal(policy)
	Expect(err).NotTo(HaveOccurred())

	return string(bytes)
}

func generateDynamicScaleInPolicy(instanceMin, instanceMax int, metricName string, threshold int64) string {
	scalingInRule := ScalingRule{
		MetricType:            metricName,
		StatWindowSeconds:     interval,
		BreachDurationSeconds: interval,
		Threshold:             threshold,
		Operator:              "<",
		CoolDownSeconds:       interval,
		Adjustment:            "-1",
	}

	policy := ScalingPolicy{
		InstanceMin:  instanceMin,
		InstanceMax:  instanceMax,
		ScalingRules: []*ScalingRule{&scalingInRule},
	}
	bytes, err := json.Marshal(policy)
	Expect(err).NotTo(HaveOccurred())

	return string(bytes)
}

func generateDynamicAndSpecificDateSchedulePolicy(instanceMin, instanceMax int, threshold int64,
	timezone string, startDateTime, endDateTime time.Time,
	scheduledInstanceMin, scheduledInstanceMax, scheduledInstanceInit int) string {

	scalingInRule := ScalingRule{
		MetricType:            "throughput",
		StatWindowSeconds:     interval,
		BreachDurationSeconds: interval,
		Threshold:             threshold,
		Operator:              "<",
		CoolDownSeconds:       interval,
		Adjustment:            "-1",
	}

	specificDateSchedule := SpecificDateSchedule{
		StartDateTime:         startDateTime.Format("2006-01-02T15:04"),
		EndDateTime:           endDateTime.Format("2006-01-02T15:04"),
		ScheduledInstanceMin:  scheduledInstanceMin,
		ScheduledInstanceMax:  scheduledInstanceMax,
		ScheduledInstanceInit: scheduledInstanceInit,
	}

	policy := ScalingPolicy{
		InstanceMin:  instanceMin,
		InstanceMax:  instanceMax,
		ScalingRules: []*ScalingRule{&scalingInRule},
		Schedules: &ScalingSchedules{
			Timezone:              timezone,
			SpecificDateSchedules: []*SpecificDateSchedule{&specificDateSchedule},
		},
	}

	bytes, err := json.Marshal(policy)
	Expect(err).NotTo(HaveOccurred())

	return string(bytes)
}

func generateDynamicAndRecurringSchedulePolicy(instanceMin, instanceMax int, threshold int64,
	timezone string, startTime, endTime time.Time, daysOfMonthOrWeek Days,
	scheduledInstanceMin, scheduledInstanceMax, scheduledInstanceInit int) string {

	scalingInRule := ScalingRule{
		MetricType:            "throughput",
		StatWindowSeconds:     interval,
		BreachDurationSeconds: interval,
		Threshold:             threshold,
		Operator:              "<",
		CoolDownSeconds:       interval,
		Adjustment:            "-1",
	}

	recurringSchedule := RecurringSchedule{
		StartTime:             startTime.Format("15:04"),
		EndTime:               endTime.Format("15:04"),
		ScheduledInstanceMin:  scheduledInstanceMin,
		ScheduledInstanceMax:  scheduledInstanceMax,
		ScheduledInstanceInit: scheduledInstanceInit,
	}

	if daysOfMonthOrWeek == daysOfMonth {
		day := startTime.Day()
		recurringSchedule.DaysOfMonth = []int{day}

	} else {
		day := int(startTime.Weekday())
		if day == 0 {
			day = 7
		}
		recurringSchedule.DaysOfWeek = []int{day}
	}

	policy := ScalingPolicy{
		InstanceMin:  instanceMin,
		InstanceMax:  instanceMax,
		ScalingRules: []*ScalingRule{&scalingInRule},
		Schedules: &ScalingSchedules{
			Timezone:           timezone,
			RecurringSchedules: []*RecurringSchedule{&recurringSchedule},
		},
	}

	bytes, err := json.Marshal(policy)
	Expect(err).NotTo(HaveOccurred())

	return string(bytes)
}

func getStartAndEndTime(location *time.Location, offset, duration time.Duration) (time.Time, time.Time) {
	// Since the validation of time could fail if spread over two days and will result in acceptance test failure
	// Need to fix dates in that case.
	startTime := time.Now().In(location).Add(offset).Truncate(time.Minute)
	if startTime.Day() != startTime.Add(duration).Day() {
		startTime = startTime.Add(duration).Truncate(24 * time.Hour)
	}
	endTime := startTime.Add(duration)
	return startTime, endTime
}
