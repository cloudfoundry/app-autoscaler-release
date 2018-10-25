package helpers

import (
	"acceptance/config"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

const (
	DaysOfMonth Days = "days_of_month"
	DaysOfWeek       = "days_of_week"
	MB               = 1024 * 1024

	TestBreachDurationSeconds = 60
	TestCoolDownSeconds       = 60
)

type appSummary struct {
	RunningInstances int `json:"running_instances"`
}

type instanceStats struct {
	MemQuota uint64 `json:"mem_quota"`
	Usage    instanceUsage
}

type instanceUsage struct {
	Mem uint64
	Cpu float64
}

type instanceInfo struct {
	State string
	Stats instanceStats
}

type appStats map[string]*instanceInfo

type Days string

type ScalingPolicy struct {
	InstanceMin  int               `json:"instance_min_count"`
	InstanceMax  int               `json:"instance_max_count"`
	ScalingRules []*ScalingRule    `json:"scaling_rules,omitempty"`
	Schedules    *ScalingSchedules `json:"schedules,omitempty"`
}

type ScalingRule struct {
	MetricType            string `json:"metric_type"`
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

func Curl(cfg *config.Config, args ...string) (int, []byte, error) {
	curlCmd := helpers.Curl(cfg, append([]string{"--output", "/dev/stderr", "--write-out", "%{http_code}"}, args...)...).Wait(cfg.DefaultTimeoutDuration())
	if curlCmd.ExitCode() != 0 {
		return 0, curlCmd.Err.Contents(), fmt.Errorf("curl failed: exit code %d", curlCmd.ExitCode())
	}
	statusCode, err := strconv.Atoi(string(curlCmd.Out.Contents()))
	if err != nil {
		return 0, curlCmd.Err.Contents(), err
	}
	return statusCode, curlCmd.Err.Contents(), nil
}

func OauthToken(cfg *config.Config) string {
	cmd := cf.Cf("oauth-token")
	Expect(cmd.Wait(cfg.DefaultTimeoutDuration())).To(Exit(0))
	return strings.TrimSpace(string(cmd.Out.Contents()))
}

func EnableServiceAccess(cfg *config.Config, orgName string) {
	enableServiceAccess := cf.Cf("enable-service-access", cfg.ServiceName, "-o", orgName).Wait(cfg.DefaultTimeoutDuration())
	Expect(enableServiceAccess).To(Exit(0), fmt.Sprintf("Failed to enable service %s for org %s", cfg.ServiceName, orgName))
}

func DisableServiceAccess(cfg *config.Config, orgName string) {
	enableServiceAccess := cf.Cf("disable-service-access", cfg.ServiceName, "-o", orgName).Wait(cfg.DefaultTimeoutDuration())
	Expect(enableServiceAccess).To(Exit(0), fmt.Sprintf("Failed to disable service %s for org %s", cfg.ServiceName, orgName))
}

func GenerateDynamicScaleOutPolicy(cfg *config.Config, instanceMin, instanceMax int, metricName string, threshold int64) string {
	scalingOutRule := ScalingRule{
		MetricType:            metricName,
		BreachDurationSeconds: TestBreachDurationSeconds,
		Threshold:             threshold,
		Operator:              ">=",
		CoolDownSeconds:       TestCoolDownSeconds,
		Adjustment:            "+1",
	}

	policy := ScalingPolicy{
		InstanceMin:  instanceMin,
		InstanceMax:  instanceMax,
		ScalingRules: []*ScalingRule{&scalingOutRule},
	}
	bytes, err := MarshalWithoutHTMLEscape(policy)
	Expect(err).NotTo(HaveOccurred())

	return string(bytes)
}

func GenerateDynamicScaleInPolicy(cfg *config.Config, instanceMin, instanceMax int, metricName string, threshold int64) string {
	scalingInRule := ScalingRule{
		MetricType:            metricName,
		BreachDurationSeconds: TestBreachDurationSeconds,
		Threshold:             threshold,
		Operator:              "<",
		CoolDownSeconds:       TestCoolDownSeconds,
		Adjustment:            "-1",
	}

	policy := ScalingPolicy{
		InstanceMin:  instanceMin,
		InstanceMax:  instanceMax,
		ScalingRules: []*ScalingRule{&scalingInRule},
	}
	bytes, err := MarshalWithoutHTMLEscape(policy)
	Expect(err).NotTo(HaveOccurred())

	return string(bytes)
}

func GenerateDynamicAndSpecificDateSchedulePolicy(cfg *config.Config, instanceMin, instanceMax int, threshold int64,
	timezone string, startDateTime, endDateTime time.Time,
	scheduledInstanceMin, scheduledInstanceMax, scheduledInstanceInit int) string {

	scalingInRule := ScalingRule{
		MetricType:            "memoryused",
		BreachDurationSeconds: TestBreachDurationSeconds,
		Threshold:             threshold,
		Operator:              "<",
		CoolDownSeconds:       TestCoolDownSeconds,
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

	bytes, err := MarshalWithoutHTMLEscape(policy)
	Expect(err).NotTo(HaveOccurred())

	return string(bytes)
}

func GenerateDynamicAndRecurringSchedulePolicy(cfg *config.Config, instanceMin, instanceMax int, threshold int64,
	timezone string, startTime, endTime time.Time, daysOfMonthOrWeek Days,
	scheduledInstanceMin, scheduledInstanceMax, scheduledInstanceInit int) string {

	scalingInRule := ScalingRule{
		MetricType:            "memoryused",
		BreachDurationSeconds: TestBreachDurationSeconds,
		Threshold:             threshold,
		Operator:              "<",
		CoolDownSeconds:       TestCoolDownSeconds,
		Adjustment:            "-1",
	}

	recurringSchedule := RecurringSchedule{
		StartTime:             startTime.Format("15:04"),
		EndTime:               endTime.Format("15:04"),
		ScheduledInstanceMin:  scheduledInstanceMin,
		ScheduledInstanceMax:  scheduledInstanceMax,
		ScheduledInstanceInit: scheduledInstanceInit,
	}

	if daysOfMonthOrWeek == DaysOfMonth {
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

	bytes, err := MarshalWithoutHTMLEscape(policy)
	Expect(err).NotTo(HaveOccurred())

	return string(bytes)
}

func RunningInstances(appGUID string, timeout time.Duration) int {
	cmd := cf.Cf("curl", "/v2/apps/"+appGUID+"/summary")
	Expect(cmd.Wait(timeout)).To(Exit(0))

	var summary appSummary
	err := json.Unmarshal(cmd.Out.Contents(), &summary)
	Expect(err).ToNot(HaveOccurred())
	return summary.RunningInstances
}

func WaitForNInstancesRunning(appGUID string, instances int, timeout time.Duration) {
	Eventually(func() int {
		return RunningInstances(appGUID, timeout)
	}, timeout, 10*time.Second).Should(Equal(instances))
}

func allInstancesStatsUsed(appGUID string, timeout time.Duration) ([]uint64, []float64) {
	cmd := cf.Cf("curl", "/v2/apps/"+appGUID+"/stats")
	Expect(cmd.Wait(timeout)).To(Exit(0))

	var stats appStats
	err := json.Unmarshal(cmd.Out.Contents(), &stats)
	Expect(err).ToNot(HaveOccurred())

	if len(stats) == 0 {
		return []uint64{}, []float64{}
	}

	mem := make([]uint64, len(stats))
	cpu := make([]float64, len(stats))

	for k, instance := range stats {
		i, err := strconv.Atoi(k)
		Expect(err).NotTo(HaveOccurred())
		mem[i] = instance.Stats.Usage.Mem
		cpu[i] = instance.Stats.Usage.Cpu
	}
	return mem, cpu
}

func AverageStatsUsedByInstance(appGUID string, timeout time.Duration) (uint64, float64) {
	memoryUsedArray, cpuUsedArray := allInstancesStatsUsed(appGUID, timeout)
	instanceCount := len(memoryUsedArray)
	if instanceCount == 0 {
		return math.MaxInt64, math.MaxFloat64
	}

	var memSum uint64
	for _, m := range memoryUsedArray {
		memSum += m
	}

	var cpuSum float64
	for _, m := range cpuUsedArray {
		cpuSum += m
	}
	avgMem := memSum / uint64(len(memoryUsedArray))
	avgCpu := cpuSum / float64(len(cpuUsedArray))
	return avgMem, avgCpu
}

func MarshalWithoutHTMLEscape(v interface{}) ([]byte, error) {

	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil

}
