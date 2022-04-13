package helpers

import (
	"acceptance/config"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/KevinJCross/cf-test-helpers/v2/generator"

	"github.com/KevinJCross/cf-test-helpers/v2/cf"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

const (
	DaysOfMonth Days = "days_of_month"
	DaysOfWeek       = "days_of_week"
	MB               = 1024 * 1024

	TestBreachDurationSeconds = 60
	TestCoolDownSeconds       = 60

	PolicyPath = "/v1/apps/{appId}/policy"
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
	CPU float64
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

type ScalingPolicyWithExtraFields struct {
	IsAdmin      bool                           `json:"is_admin"`
	IsSSO        bool                           `json:"is_sso"`
	Role         string                         `json:"role"`
	InstanceMin  int                            `json:"instance_min_count"`
	InstanceMax  int                            `json:"instance_max_count"`
	ScalingRules []*ScalingRulesWithExtraFields `json:"scaling_rules,omitempty"`
	Schedules    *ScalingSchedules              `json:"schedules,omitempty"`
}

type ScalingRule struct {
	MetricType            string `json:"metric_type"`
	BreachDurationSeconds int    `json:"breach_duration_secs"`
	Threshold             int64  `json:"threshold"`
	Operator              string `json:"operator"`
	CoolDownSeconds       int    `json:"cool_down_secs"`
	Adjustment            string `json:"adjustment"`
}

type ScalingRulesWithExtraFields struct {
	StatsWindowSeconds int `json:"stats_window_secs"`
	ScalingRule
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

func OauthToken(cfg *config.Config) string {
	cmd := cf.Cf("oauth-token")
	Expect(cmd.Wait(cfg.DefaultTimeoutDuration())).To(Exit(0))
	return strings.TrimSpace(string(cmd.Out.Contents()))
}

func EnableServiceAccess(cfg *config.Config, orgName string) {
	enableServiceAccess := cf.Cf("enable-service-access", cfg.ServiceName, "-b", cfg.ServiceBroker, "-o", orgName).Wait(cfg.DefaultTimeoutDuration())
	Expect(enableServiceAccess).To(Exit(0), fmt.Sprintf("Failed to enable service %s for org %s", cfg.ServiceName, orgName))
}

func DisableServiceAccess(cfg *config.Config, orgName string) {
	enableServiceAccess := cf.Cf("disable-service-access", cfg.ServiceName, "-b", cfg.ServiceBroker, "-o", orgName).Wait(cfg.DefaultTimeoutDuration())
	Expect(enableServiceAccess).To(Exit(0), fmt.Sprintf("Failed to disable service %s for org %s", cfg.ServiceName, orgName))
}

func CheckServiceExists(cfg *config.Config) {
	serviceExists := cf.Cf("marketplace", "-e", cfg.ServiceName).Wait(cfg.DefaultTimeoutDuration())

	Expect(serviceExists).To(Exit(0), fmt.Sprintf("Service offering, %s, does not exist", cfg.ServiceName))
}

func GenerateDynamicScaleOutPolicy(instanceMin, instanceMax int, metricName string, threshold int64) string {
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

func GenerateDynamicScaleOutPolicyWithExtraFields(instanceMin, instanceMax int, metricName string, threshold int64) (string, string) {
	scalingOutRule := ScalingRule{
		MetricType:            metricName,
		BreachDurationSeconds: TestBreachDurationSeconds,
		Threshold:             threshold,
		Operator:              ">=",
		CoolDownSeconds:       TestCoolDownSeconds,
		Adjustment:            "+1",
	}

	scalingOutRuleWithExtraFields := ScalingRulesWithExtraFields{
		StatsWindowSeconds: 666,
		ScalingRule:        scalingOutRule,
	}

	policy := ScalingPolicy{
		InstanceMin:  instanceMin,
		InstanceMax:  instanceMax,
		ScalingRules: []*ScalingRule{&scalingOutRule},
	}

	policyWithExtraFields := ScalingPolicyWithExtraFields{
		IsAdmin:      true,
		IsSSO:        true,
		Role:         "admin",
		InstanceMin:  instanceMin,
		InstanceMax:  instanceMax,
		ScalingRules: []*ScalingRulesWithExtraFields{&scalingOutRuleWithExtraFields},
	}

	validBytes, err := MarshalWithoutHTMLEscape(policy)
	Expect(err).NotTo(HaveOccurred())

	extraBytes, err := MarshalWithoutHTMLEscape(policyWithExtraFields)
	Expect(err).NotTo(HaveOccurred())

	Expect(extraBytes).NotTo(MatchJSON(validBytes))

	return string(extraBytes), string(validBytes)
}

func GenerateDynamicScaleInPolicy(instanceMin, instanceMax int, metricName string, threshold int64) string {
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

func GenerateDynamicScaleOutAndInPolicy(instanceMin, instanceMax int, metricName string, minThreshold int64, maxThreshold int64) string {
	scalingInRule := ScalingRule{
		MetricType:            metricName,
		BreachDurationSeconds: TestBreachDurationSeconds,
		Threshold:             minThreshold,
		Operator:              "<",
		CoolDownSeconds:       TestCoolDownSeconds,
		Adjustment:            "-1",
	}

	scalingOutRule := ScalingRule{
		MetricType:            metricName,
		BreachDurationSeconds: TestBreachDurationSeconds,
		Threshold:             maxThreshold,
		Operator:              ">=",
		CoolDownSeconds:       TestCoolDownSeconds,
		Adjustment:            "+1",
	}

	policy := ScalingPolicy{
		InstanceMin:  instanceMin,
		InstanceMax:  instanceMax,
		ScalingRules: []*ScalingRule{&scalingOutRule, &scalingInRule},
	}

	bytes, err := MarshalWithoutHTMLEscape(policy)
	Expect(err).NotTo(HaveOccurred())

	return string(bytes)
}

func GenerateDynamicAndSpecificDateSchedulePolicy(instanceMin, instanceMax int, threshold int64,
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

	return strings.TrimSpace(string(bytes))
}

func GenerateDynamicAndRecurringSchedulePolicy(instanceMin, instanceMax int, threshold int64,
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

func AverageMemoryUsedByInstance(appGUID string, timeout time.Duration) uint64 {
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

func CreatePolicy(cfg *config.Config, appName, appGUID, policy string) string {
	if cfg.IsServiceOfferingEnabled() {
		instanceName := generator.PrefixedRandomName(cfg.Prefix, cfg.InstancePrefix)
		createService := cf.Cf("create-service", cfg.ServiceName, cfg.ServicePlan, instanceName, "-b", cfg.ServiceBroker).Wait(cfg.DefaultTimeoutDuration())
		Expect(createService).To(Exit(0), "failed creating service")

		bindService := cf.Cf("bind-service", appName, instanceName, "-c", policy).Wait(cfg.DefaultTimeoutDuration())
		Expect(bindService).To(Exit(0), "failed binding service to app with a policy ")
		return instanceName
	}
	CreatePolicyWithAPI(cfg, appGUID, policy)
	return ""
}

func CreatePolicyWithAPI(cfg *config.Config, appGUID, policy string) {
	oauthToken := OauthToken(cfg)
	client := GetHTTPClient(cfg)

	policyURL := fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(PolicyPath, "{appId}", appGUID, -1))
	req, err := http.NewRequest("PUT", policyURL, bytes.NewBuffer([]byte(policy)))
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	Expect(err).ShouldNot(HaveOccurred())
	defer resp.Body.Close()
	Expect(resp.StatusCode == 200 || resp.StatusCode == 201).Should(BeTrue())
	Expect([]int{http.StatusOK, http.StatusCreated}).To(ContainElement(resp.StatusCode))
}

func GetHTTPClient(cfg *config.Config) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
			DisableCompression:  true,
			DisableKeepAlives:   true,
			// #nosec G402
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.SkipSSLValidation,
			},
		},
		Timeout: 30 * time.Second,
	}
}

func GetAppGuid(cfg *config.Config, appName string) string {
	guid := cf.Cf("app", appName, "--guid").Wait(cfg.DefaultTimeoutDuration())
	Expect(guid).To(Exit(0))
	return strings.TrimSpace(string(guid.Out.Contents()))
}
