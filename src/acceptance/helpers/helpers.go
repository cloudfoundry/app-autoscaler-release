package helpers

import (
	"acceptance/config"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	url2 "net/url"
	"strings"
	"time"

	"github.com/KevinJCross/cf-test-helpers/v2/workflowhelpers"

	. "github.com/onsi/ginkgo/v2"

	"github.com/KevinJCross/cf-test-helpers/v2/generator"

	"github.com/KevinJCross/cf-test-helpers/v2/cf"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

const (
	DaysOfMonth Days = "days_of_month"
	DaysOfWeek  Days = "days_of_week"

	TestBreachDurationSeconds = 60
	TestCoolDownSeconds       = 60

	PolicyPath = "/v1/apps/{appId}/policy"
)

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
	cmd := cf.CfSilent("oauth-token")
	Expect(cmd.Wait(cfg.DefaultTimeoutDuration())).To(Exit(0))
	return strings.TrimSpace(string(cmd.Out.Contents()))
}

func EnableServiceAccess(setup *workflowhelpers.ReproducibleTestSuiteSetup, cfg *config.Config, orgName string) {
	if cfg.IsServiceOfferingEnabled() && cfg.ShouldEnableServiceAccess() {
		workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
			if orgName == "" {
				Fail(fmt.Sprintf("Org must not be an empty string. Using broker:%s, serviceName:%s", cfg.ServiceBroker, cfg.ServiceName))
			}
			enableServiceAccess := cf.Cf("enable-service-access", cfg.ServiceName, "-b", cfg.ServiceBroker, "-o", orgName).Wait(cfg.DefaultTimeoutDuration())
			Expect(enableServiceAccess).To(Exit(0), fmt.Sprintf("Failed to enable service %s for org %s", cfg.ServiceName, orgName))
		})
	}
}

func DisableServiceAccess(cfg *config.Config, setup *workflowhelpers.ReproducibleTestSuiteSetup) {
	if cfg.IsServiceOfferingEnabled() && cfg.ShouldEnableServiceAccess() {
		workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
			orgName := setup.GetOrganizationName()
			enableServiceAccess := cf.Cf("disable-service-access", cfg.ServiceName, "-b", cfg.ServiceBroker, "-o", orgName).Wait(cfg.DefaultTimeoutDuration())
			Expect(enableServiceAccess).To(Exit(0), fmt.Sprintf("Failed to disable service %s for org %s", cfg.ServiceName, orgName))
		})
	}
}

func CheckServiceExists(cfg *config.Config, spaceName, serviceName string) {
	if cfg.IsServiceOfferingEnabled() {
		spaceCmd := cf.Cf("space", spaceName, "--guid").Wait(cfg.DefaultTimeoutDuration())
		Expect(spaceCmd).To(Exit(0), fmt.Sprintf("Space, %s, does not exist", spaceName))
		spaceGuid := strings.TrimSpace(strings.Trim(string(spaceCmd.Out.Contents()), "\n"))

		serviceCmd := cf.CfSilent("curl", "-f", ServicePlansUrl(cfg, spaceGuid)).Wait(cfg.DefaultTimeoutDuration())
		if serviceCmd.ExitCode() != 0 {
			Fail(fmt.Sprintf("Failed get broker information for serviceName=%s spaceName=%s", cfg.ServiceName, spaceName))
		}

		var services = struct {
			Included struct {
				ServiceOfferings []struct{ Name string } `json:"service_offerings"`
			}
		}{}
		contents := serviceCmd.Out.Contents()
		err := json.Unmarshal(contents, &services)
		if err != nil {
			AbortSuite(fmt.Sprintf("Failed to parse service plan json: %s\n\n'%s'", err.Error(), string(contents)))
		}
		GinkgoWriter.Printf("\nFound services: %s\n", services.Included.ServiceOfferings)
		for _, service := range services.Included.ServiceOfferings {
			if service.Name == serviceName {
				return
			}
		}

		cf.Cf("marketplace", "-e", cfg.ServiceName).Wait(cfg.DefaultTimeoutDuration())
		Fail(fmt.Sprintf("Could not find service %s in space %s", serviceName, spaceName))
	}
}

func ServicePlansUrl(cfg *config.Config, spaceGuid string) string {
	values := url2.Values{
		"available": []string{"true"},
		"fields[service_offering.service_broker]": []string{"name,guid"},
		"include":                []string{"service_offering"},
		"per_page":               []string{"5000"},
		"service_broker_names":   []string{cfg.ServiceBroker},
		"service_offering_names": []string{cfg.ServiceName},
		"space_guids":            []string{spaceGuid},
	}
	url := &url2.URL{Path: "/v3/service_plans", RawQuery: values.Encode()}
	return url.String()
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
	marshaled, err := MarshalWithoutHTMLEscape(policy)
	Expect(err).NotTo(HaveOccurred())

	return string(marshaled)
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
	marshaled, err := MarshalWithoutHTMLEscape(policy)
	Expect(err).NotTo(HaveOccurred())

	return string(marshaled)
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

	marshaled, err := MarshalWithoutHTMLEscape(policy)
	Expect(err).NotTo(HaveOccurred())

	return string(marshaled)
}

func GenerateSpecificDateSchedulePolicy(startDateTime, endDateTime time.Time, scheduledInstanceMin, scheduledInstanceMax, scheduledInstanceInit int) string {
	scalingInRule := ScalingRule{
		MetricType:            "cpu",
		BreachDurationSeconds: TestBreachDurationSeconds,
		Threshold:             80,
		Operator:              "<",
		CoolDownSeconds:       TestCoolDownSeconds,
		Adjustment:            "-1",
	}
	specificDateSchedule := SpecificDateSchedule{
		StartDateTime:         startDateTime.Round(1 * time.Minute).Format("2006-01-02T15:04"),
		EndDateTime:           endDateTime.Round(1 * time.Minute).Format("2006-01-02T15:04"),
		ScheduledInstanceMin:  scheduledInstanceMin,
		ScheduledInstanceMax:  scheduledInstanceMax,
		ScheduledInstanceInit: scheduledInstanceInit,
	}
	policy := ScalingPolicy{
		InstanceMin:  1,
		InstanceMax:  4,
		ScalingRules: []*ScalingRule{&scalingInRule},
		Schedules: &ScalingSchedules{
			Timezone:              "UTC",
			SpecificDateSchedules: []*SpecificDateSchedule{&specificDateSchedule},
		},
	}

	marshaled, err := MarshalWithoutHTMLEscape(policy)
	Expect(err).NotTo(HaveOccurred())

	return strings.TrimSpace(string(marshaled))
}

func GenerateDynamicAndRecurringSchedulePolicy(instanceMin, instanceMax int, threshold int64,
	timezone string, startTime, endTime time.Time, daysOfMonthOrWeek Days,
	scheduledInstanceMin, scheduledInstanceMax, scheduledInstanceInit int) string {
	scalingInRule := ScalingRule{
		MetricType:            "cpu",
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

	marshaled, err := MarshalWithoutHTMLEscape(policy)
	Expect(err).NotTo(HaveOccurred())

	return string(marshaled)
}

func RunningInstances(appGUID string, timeout time.Duration) int {
	defer GinkgoRecover()
	cmd := cf.CfSilent("curl", fmt.Sprintf("/v3/apps/%s/processes/web", appGUID)).Wait(timeout)
	Expect(cmd).To(Exit(0))
	var process = struct {
		Instances int `json:"instances"`
	}{}

	err := json.Unmarshal(cmd.Out.Contents(), &process)
	Expect(err).ToNot(HaveOccurred())
	webInstances := process.Instances
	GinkgoWriter.Printf("\nFound %d app instances\n", webInstances)
	return webInstances
}

func WaitForNInstancesRunning(appGUID string, instances int, timeout time.Duration) {
	By(fmt.Sprintf("Waiting for %d instances of app: %s", instances, appGUID))
	Eventually(getAppInstances(appGUID, 8*time.Second)).
		WithOffset(1).
		WithTimeout(timeout).
		WithPolling(10 * time.Second).
		Should(Equal(instances))
}

func getAppInstances(appGUID string, timeout time.Duration) func() int {
	return func() int { return RunningInstances(appGUID, timeout) }
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
		instanceName := CreateService(cfg)
		BindServiceToAppWithPolicy(cfg, appName, instanceName, policy)
		return instanceName
	}
	CreatePolicyWithAPI(cfg, appGUID, policy)
	return ""
}

func BindServiceToApp(cfg *config.Config, appName string, instanceName string) {
	BindServiceToAppWithPolicy(cfg, appName, instanceName, "")
}

func BindServiceToAppWithPolicy(cfg *config.Config, appName string, instanceName string, policy string) {
	if cfg.IsServiceOfferingEnabled() {
		args := []string{"bind-service", appName, instanceName}
		if policy != "" {
			args = append(args, "-c", policy)
		}
		bindService := cf.Cf(args...).Wait(cfg.DefaultTimeoutDuration())
		FailOnCommandFailuref(bindService, "failed binding service %s to app %s. \n Command Error: %s %s", instanceName, appName, bindService.Buffer().Contents(), bindService.Err.Contents())
	}
}

func CreateService(cfg *config.Config) string {
	return CreateServiceWithPlan(cfg, cfg.ServicePlan)
}

func CreateServiceWithPlan(cfg *config.Config, servicePlan string) string {
	return CreateServiceWithPlanAndParameters(cfg, servicePlan, "")
}

func CreateServiceWithPlanAndParameters(cfg *config.Config, servicePlan string, defaultPolicy string) string {
	if cfg.IsServiceOfferingEnabled() {
		instanceName := generator.PrefixedRandomName(cfg.Prefix, cfg.InstancePrefix)
		cfCommand := []string{"create-service", cfg.ServiceName, servicePlan, instanceName, "-b", cfg.ServiceBroker}
		if defaultPolicy != "" {
			cfCommand = append(cfCommand, "-c", defaultPolicy)
		}
		createService := cf.Cf(cfCommand...).Wait(cfg.DefaultTimeoutDuration())
		FailOnCommandFailuref(createService, "Failed to create service instance %s on service %s \n Command Error: %s %s", instanceName, cfg.ServiceName, createService.Buffer().Contents(), createService.Err.Contents())
		return instanceName
	}
	return ""
}

func GetServiceInstanceGuid(cfg *config.Config, instanceName string) string {
	guid := cf.Cf("service", instanceName, "--guid").Wait(cfg.DefaultTimeoutDuration())
	Expect(guid).To(Exit(0), fmt.Sprintf("Failed to find service instance guid for service instance: %s \n CLI Output:\n %s", instanceName, guid.Out.Contents()))
	return strings.TrimSpace(string(guid.Out.Contents()))
}

func GetServiceInstanceParameters(cfg *config.Config, instanceName string) string {
	instanceGuid := GetServiceInstanceGuid(cfg, instanceName)

	cmd := cf.CfSilent("curl", fmt.Sprintf("/v3/service_instances/%s/parameters", instanceGuid)).Wait(cfg.DefaultTimeoutDuration())
	Expect(cmd).To(Exit(0))
	return strings.TrimSpace(string(cmd.Out.Contents()))
}

func GetServiceCredentialBindingGuid(cfg *config.Config, instanceGuid string, appName string) string {
	appGuid := GetAppGuid(cfg, appName)
	guid := cf.CfSilent("curl", fmt.Sprintf("/v3/service_credential_bindings?service_instance_guids=%s&app_guids=%s", instanceGuid, appGuid)).Wait(cfg.DefaultTimeoutDuration())

	Expect(guid).To(Exit(0), fmt.Sprintf("Failed to find service credential binding guid for service instance guid : %s and app name %s \n CLI Output:\n %s", instanceGuid, appName, guid.Out.Contents()))

	contents := guid.Out.Contents()

	type ServiceCredentialBinding struct {
		GUID string `json:"guid"`
	}

	var serviceCredentialBindings = struct {
		Resources []ServiceCredentialBinding `json:"resources"`
	}{}
	err := json.Unmarshal(contents, &serviceCredentialBindings)
	Expect(err).ShouldNot(HaveOccurred())

	return serviceCredentialBindings.Resources[0].GUID
}

func GetServiceCredentialBindingParameters(cfg *config.Config, instanceName string, appName string) string {
	instanceGuid := GetServiceInstanceGuid(cfg, instanceName)
	serviceCredentialBindingGuid := GetServiceCredentialBindingGuid(cfg, instanceGuid, appName)

	cmd := cf.CfSilent("curl", fmt.Sprintf("/v3/service_credential_bindings/%s/parameters", serviceCredentialBindingGuid)).Wait(cfg.DefaultTimeoutDuration())
	Expect(cmd).To(Exit(0))
	return strings.TrimSpace(string(cmd.Out.Contents()))
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
	defer func() { _ = resp.Body.Close() }()
	Expect(resp.StatusCode == 200 || resp.StatusCode == 201).Should(BeTrue())
	Expect([]int{http.StatusOK, http.StatusCreated}).To(ContainElement(resp.StatusCode))
}

func GetHTTPClient(cfg *config.Config) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
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
	Expect(guid).To(Exit(0), fmt.Sprintf("Failed to find app guid for app: %s \n CLI Output:\n %s", appName, guid.Out.Contents()))
	return strings.TrimSpace(string(guid.Out.Contents()))
}

func FailOnCommandFailuref(command *Session, format string, args ...any) *Session {
	if command.ExitCode() != 0 {
		Fail(fmt.Sprintf(format, args...))
	}
	return command
}
