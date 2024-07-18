package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/cf"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/cf/mocks"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/db"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"
	. "code.cloudfoundry.org/app-autoscaler/src/autoscaler/testhelpers"

	"code.cloudfoundry.org/lager/v3"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon_v2"
	"github.com/tedsuo/ifrit/grouper"
)

const (
	serviceId     = "autoscaler-guid"
	planId        = "autoscaler-free-plan-id"
	testUserId    = "testUserId"
	testUserToken = "testUserOauthToken" // #nosec G101
)

var (
	components              Components
	tmpDir                  string
	golangApiServerConfPath string
	schedulerConfPath       string
	eventGeneratorConfPath  string
	scalingEngineConfPath   string
	operatorConfPath        string
	brokerAuth              string
	dbUrl                   string
	LOGLEVEL                string
	dbHelper                *sqlx.DB
	fakeCCNOAAUAA           *mocks.Server
	testUserScope           = []string{"cloud_controller.read", "cloud_controller.write", "password.write", "openid", "network.admin", "network.write", "uaa.user"}
	processMap              = map[string]ifrit.Process{}
	mockLogCache            = &MockLogCache{}

	defaultHttpClientTimeout = 10 * time.Second

	saveInterval              = 1 * time.Second
	aggregatorExecuteInterval = 1 * time.Second
	policyPollerInterval      = 1 * time.Second
	evaluationManagerInterval = 1 * time.Second
	breachDurationSecs        = 5

	httpClient             *http.Client
	httpClientForPublicApi *http.Client
	logger                 lager.Logger

	testCertDir = "../../../test-certs"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	components = Components{
		Ports:       PreparePorts(),
		Executables: CompileTestedExecutables(),
	}
	payload, err := json.Marshal(&components)
	Expect(err).NotTo(HaveOccurred())

	dbUrl = GetDbUrl()
	if dbUrl == "" {
		Fail("environment variable $DBURL is not set")
	}

	database, err := db.GetConnection(dbUrl)
	Expect(err).NotTo(HaveOccurred())

	dbHelper, err = sqlx.Open(database.DriverName, database.DSN)
	Expect(err).NotTo(HaveOccurred())

	clearDatabase()

	return payload
}, func(encodedBuiltArtifacts []byte) {
	err := json.Unmarshal(encodedBuiltArtifacts, &components)
	Expect(err).NotTo(HaveOccurred())
	components.Ports = PreparePorts()

	tmpDir, err = os.MkdirTemp("", "autoscaler")
	Expect(err).NotTo(HaveOccurred())

	dbUrl = GetDbUrl()
	database, err := db.GetConnection(dbUrl)
	Expect(err).NotTo(HaveOccurred())

	dbHelper, err = sqlx.Open(database.DriverName, database.DSN)
	Expect(err).NotTo(HaveOccurred())

	LOGLEVEL = os.Getenv("LOGLEVEL")
	if LOGLEVEL == "" {
		LOGLEVEL = "info"
	}
})

var _ = SynchronizedAfterSuite(func() {
	if len(tmpDir) > 0 {
		_ = os.RemoveAll(tmpDir)
	}
}, func() {

})

var _ = BeforeEach(func() {
	httpClient = NewApiClient()
	httpClientForPublicApi = NewPublicApiClient()
	logger = lager.NewLogger("test")
	logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
})

func CompileTestedExecutables() Executables {
	builtExecutables := Executables{}
	var err error
	workingDir, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	rootDir := path.Join(workingDir, "..", "..", "..")

	builtExecutables[Scheduler] = path.Join(rootDir, "src", "scheduler", "target", "scheduler-1.0-SNAPSHOT.war")
	builtExecutables[EventGenerator] = path.Join(rootDir, "src", "autoscaler", "build", "eventgenerator")
	builtExecutables[ScalingEngine] = path.Join(rootDir, "src", "autoscaler", "build", "scalingengine")
	builtExecutables[Operator] = path.Join(rootDir, "src", "autoscaler", "build", "operator")
	builtExecutables[GolangAPIServer] = path.Join(rootDir, "src", "autoscaler", "build", "api")

	return builtExecutables
}

func PreparePorts() Ports {
	return Ports{
		GolangAPIServer:     22000 + GinkgoParallelProcess(),
		GolangServiceBroker: 23000 + GinkgoParallelProcess(),
		Scheduler:           15000 + GinkgoParallelProcess(),
		MetricsCollector:    16000 + GinkgoParallelProcess(),
		EventGenerator:      17000 + GinkgoParallelProcess(),
		ScalingEngine:       18000 + GinkgoParallelProcess(),
	}
}

func startGolangApiServer() {
	processMap[GolangAPIServer] = ginkgomon_v2.Invoke(grouper.NewOrdered(os.Interrupt, grouper.Members{
		{GolangAPIServer, components.GolangAPIServer(golangApiServerConfPath)},
	}))
}

func startScheduler() {
	processMap[Scheduler] = ginkgomon_v2.Invoke(grouper.NewOrdered(os.Interrupt, grouper.Members{
		{Scheduler, components.Scheduler(schedulerConfPath)},
	}))
}

func startEventGenerator() {
	processMap[EventGenerator] = ginkgomon_v2.Invoke(grouper.NewOrdered(os.Interrupt, grouper.Members{
		{EventGenerator, components.EventGenerator(eventGeneratorConfPath)},
	}))
}

func startScalingEngine() {
	processMap[ScalingEngine] = ginkgomon_v2.Invoke(grouper.NewOrdered(os.Interrupt, grouper.Members{
		{ScalingEngine, components.ScalingEngine(scalingEngineConfPath)},
	}))
}

func startOperator() {
	processMap[Operator] = ginkgomon_v2.Invoke(grouper.NewOrdered(os.Interrupt, grouper.Members{
		{Operator, components.Operator(operatorConfPath)},
	}))
}

func startMockLogCache() {
	tlsConfig, err := NewTLSConfig(
		filepath.Join(testCertDir, "autoscaler-ca.crt"),
		filepath.Join(testCertDir, "log-cache.crt"),
		filepath.Join(testCertDir, "log-cache.key"),
		"log-cache",
	)
	Expect(err).ToNot(HaveOccurred())

	mockLogCache = NewMockLogCache(tlsConfig)
	err = mockLogCache.Start(20000 + GinkgoParallelProcess())
	Expect(err).ToNot(HaveOccurred())
}

func stopGolangApiServer() {
	ginkgomon_v2.Kill(processMap[GolangAPIServer], 5*time.Second)
}
func stopScheduler() {
	ginkgomon_v2.Kill(processMap[Scheduler], 5*time.Second)
}
func stopScalingEngine() {
	ginkgomon_v2.Kill(processMap[ScalingEngine], 5*time.Second)
}
func stopEventGenerator() {
	ginkgomon_v2.Kill(processMap[EventGenerator], 5*time.Second)
}
func stopOperator() {
	ginkgomon_v2.Kill(processMap[Operator], 5*time.Second)
}

func stopMockLogCache() {
	mockLogCache.Stop()
}

func getRandomIdRef(ref string) string {
	report := CurrentSpecReport()
	// 0123456789012345678901234567890123456789
	// operator_others:189,11,instance:a5f63cbf 7c204c417941d91d21cb3bd0
	// |filename|:|linenumber|,|ref|process|:|random|
	// |15|1|3-4|1|14|2|1|3-4| == 40 (max id length)
	if len(ref) > 13 {
		GinkgoT().Logf("WARNING: %s:%d using a ref that is being truncated '%s' should be <= 13 chars", report.FileName(), report.LineNumber(), ref)
		ref = ref[:13]
	}
	id := fmt.Sprintf("%s:%d,%s,%d:%s", testFileFragment(report.FileName()), report.LineNumber(), ref, GinkgoParallelProcess(), randomBits())
	if len(id) > 40 {
		id = id[:40]
	}
	return id
}

func getUUID() string {
	v4, _ := uuid.NewV4()
	return v4.String()
}

func randomBits() string {
	randomBits := getUUID()
	return strings.ReplaceAll(randomBits, "-", "")
}

func testFileFragment(filename string) string {
	base := filepath.Base(filename)
	base = strings.TrimSuffix(base, "_test.go")
	base = strings.TrimPrefix(base, "integration_")
	if len(base) > 15 {
		return base[(len(base) - 15):]
	}
	return base
}

func provisionServiceInstance(serviceInstanceId string, orgId string, spaceId string, defaultPolicy []byte, serviceBrokerURL url.URL, httpClient *http.Client) (*http.Response, error) {
	By("provisionServiceInstance")
	var bindBody map[string]interface{}
	if defaultPolicy != nil {
		defaultPolicy := json.RawMessage(defaultPolicy)
		parameters := map[string]interface{}{
			"default_policy": defaultPolicy,
		}
		bindBody = map[string]interface{}{
			"organization_guid": orgId,
			"space_guid":        spaceId,
			"service_id":        serviceId,
			"plan_id":           planId,
			"parameters":        parameters,
		}
	} else {
		bindBody = map[string]interface{}{
			"organization_guid": orgId,
			"space_guid":        spaceId,
			"service_id":        serviceId,
			"plan_id":           planId,
		}
	}

	body, err := json.Marshal(bindBody)
	Expect(err).NotTo(HaveOccurred())

	serviceBrokerURL.Path = "/v2/service_instances/" + serviceInstanceId
	req, err := http.NewRequest("PUT", serviceBrokerURL.String(), bytes.NewReader(body))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+brokerAuth)
	return httpClient.Do(req)
}

func updateServiceInstance(serviceInstanceId string, defaultPolicy []byte, serviceBrokerURL url.URL, httpClient *http.Client) (*http.Response, error) {
	By("updateServiceInstance")
	var updateBody map[string]interface{}
	if defaultPolicy != nil {
		defaultPolicy := json.RawMessage(defaultPolicy)
		parameters := map[string]interface{}{
			"default_policy": &defaultPolicy,
		}
		updateBody = map[string]interface{}{
			"service_id": serviceId,
			"parameters": parameters,
		}
	}

	body, err := json.Marshal(updateBody)
	Expect(err).NotTo(HaveOccurred())

	serviceBrokerURL.Path = fmt.Sprintf("/v2/service_instances/%s", serviceInstanceId)
	req, err := http.NewRequest("PATCH", serviceBrokerURL.String(), bytes.NewReader(body))
	Expect(err).NotTo(HaveOccurred())

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+brokerAuth)
	return httpClient.Do(req)
}

func deProvisionServiceInstance(serviceInstanceId string, serviceBrokerURL url.URL, httpClient *http.Client) (*http.Response, error) {
	By("deProvisionServiceInstance")
	serviceBrokerURL.Path = fmt.Sprintf("/v2/service_instances/%s", serviceInstanceId)
	serviceBrokerURL.RawQuery = fmt.Sprintf("service_id=%s&plan_id=%s", serviceId, planId)
	req, err := http.NewRequest("DELETE", serviceBrokerURL.String(), nil)
	ExpectWithOffset(2, err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+brokerAuth)
	return httpClient.Do(req)
}

func bindService(bindingId string, appId string, serviceInstanceId string, policy []byte, serviceBrokerURL url.URL, httpClient *http.Client) (*http.Response, error) {
	By("bindService")
	var bindBody map[string]interface{}
	if policy != nil {
		rawParameters := json.RawMessage(policy)
		bindBody = map[string]interface{}{
			"app_guid":   appId,
			"service_id": serviceId,
			"plan_id":    planId,
			"parameters": rawParameters,
		}
	} else {
		bindBody = map[string]interface{}{
			"app_guid":   appId,
			"service_id": serviceId,
			"plan_id":    planId,
		}
	}

	body, err := json.Marshal(bindBody)
	Expect(err).NotTo(HaveOccurred())
	serviceBrokerURL.Path = fmt.Sprintf("/v2/service_instances/%s/service_bindings/%s", serviceInstanceId, bindingId)
	req, err := http.NewRequest("PUT", serviceBrokerURL.String(), bytes.NewReader(body))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+brokerAuth)
	return httpClient.Do(req)
}

func unbindService(bindingId string, appId string, serviceInstanceId string, serviceBrokerURL url.URL, httpClient *http.Client) (*http.Response, error) {
	By("unbindService")

	serviceBrokerURL.Path = fmt.Sprintf("/v2/service_instances/%s/service_bindings/%s", serviceInstanceId, bindingId)
	serviceBrokerURL.RawQuery = fmt.Sprintf("service_id=%s&plan_id=%s", serviceId, planId)
	req, err := http.NewRequest("DELETE", serviceBrokerURL.String(), nil)
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+brokerAuth)
	return httpClient.Do(req)
}

func provisionAndBind(serviceInstanceId string, orgId string, spaceId string, bindingId string, appId string, serviceBrokerURL url.URL, httpClient *http.Client) {
	resp, err := provisionServiceInstance(serviceInstanceId, orgId, spaceId, nil, serviceBrokerURL, httpClient)
	Expect(err).WithOffset(1).NotTo(HaveOccurred())
	Expect(resp.StatusCode).WithOffset(1).To(Equal(http.StatusCreated), fmt.Sprintf("response was '%s'", MustReadAll(resp.Body)))
	_ = resp.Body.Close()

	resp, err = bindService(bindingId, appId, serviceInstanceId, nil, serviceBrokerURL, httpClient)
	Expect(err).WithOffset(1).NotTo(HaveOccurred())
	Expect(resp.StatusCode).WithOffset(1).To(Equal(http.StatusCreated), fmt.Sprintf("response was '%s'", MustReadAll(resp.Body)))
	_ = resp.Body.Close()
}

func unbindAndDeProvision(bindingId string, appId string, serviceInstanceId string, serviceBrokerURL url.URL, httpClient *http.Client) {
	resp, err := unbindService(bindingId, appId, serviceInstanceId, serviceBrokerURL, httpClient)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, resp.StatusCode).To(Equal(http.StatusOK))
	_ = resp.Body.Close()

	resp, err = deProvisionServiceInstance(serviceInstanceId, serviceBrokerURL, httpClient)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, resp.StatusCode).To(Equal(http.StatusOK))
	_ = resp.Body.Close()
}

func getPolicy(appId string, apiURL url.URL, httpClient *http.Client) (*http.Response, error) {
	By("getPolicy")
	apiURL.Path = fmt.Sprintf("/v1/apps/%s/policy", appId)
	req, err := http.NewRequest("GET", apiURL.String(), nil)
	req.Header.Set("Authorization", "bearer fake-token")
	Expect(err).NotTo(HaveOccurred())
	return httpClient.Do(req)
}

func detachPolicy(appId string, apiURL url.URL, httpClient *http.Client) (*http.Response, error) {
	By("detachPolicy")
	apiURL.Path = fmt.Sprintf("/v1/apps/%s/policy", appId)
	req, err := http.NewRequest("DELETE", apiURL.String(), strings.NewReader(""))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "bearer fake-token")
	return httpClient.Do(req)
}

func attachPolicy(appId string, policy []byte, apiURL url.URL, httpClient *http.Client) (*http.Response, error) {
	By("attachPolicy")
	apiURL.Path = fmt.Sprintf("/v1/apps/%s/policy", appId)
	req, err := http.NewRequest("PUT", apiURL.String(), bytes.NewReader(policy))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "bearer fake-token")
	return httpClient.Do(req)
}

func getSchedules(schedulerURL url.URL, appId string) (*http.Response, error) {
	By("getSchedules")
	schedulerURL.Path = fmt.Sprintf("/v1/apps/%s/schedules", appId)
	req, err := http.NewRequest("GET", schedulerURL.String(), strings.NewReader(""))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	return httpClient.Do(req)
}

func createSchedule(appId string, guid string, schedule string, schedulerURL url.URL) (*http.Response, error) {
	By("createSchedule")
	schedulerURL.Path = fmt.Sprintf("/v1/apps/%s/schedules", appId)
	schedulerURL.RawQuery = fmt.Sprintf("guid=%s", guid)
	req, err := http.NewRequest("PUT", schedulerURL.String(), bytes.NewReader([]byte(schedule)))
	if err != nil {
		panic(err)
	}
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	return httpClient.Do(req)
}

func deleteSchedule(schedulerURL url.URL, appId string) (*http.Response, error) {
	By("deleteSchedule")
	schedulerURL.Path = fmt.Sprintf("/v1/apps/%s/schedules", appId)
	req, err := http.NewRequest("DELETE", schedulerURL.String(), strings.NewReader(""))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	return httpClient.Do(req)
}

func getActiveSchedule(scalingEngineURL url.URL, appId string) (*http.Response, error) {
	By("getActiveSchedule")
	scalingEngineURL.Path = fmt.Sprintf("/v1/apps/%s/active_schedules", appId)
	req, err := http.NewRequest("GET", scalingEngineURL.String(), strings.NewReader(""))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	return httpClient.Do(req)
}

func activeScheduleExists(scalingEngineURL url.URL, appId string) bool {
	resp, err := getActiveSchedule(scalingEngineURL, appId)
	if err == nil {
		defer func() { _ = resp.Body.Close() }()
	}
	Expect(err).NotTo(HaveOccurred())

	return resp.StatusCode == http.StatusOK
}

func setPolicyRecurringDate(policyByte []byte) []byte {
	var policy models.ScalingPolicy
	err := json.Unmarshal(policyByte, &policy)
	Expect(err).NotTo(HaveOccurred())

	if policy.Schedules != nil {
		location, err := time.LoadLocation(policy.Schedules.Timezone)
		Expect(err).NotTo(HaveOccurred())
		now := time.Now().In(location)
		starttime := now.Add(time.Minute * 10)
		endtime := now.Add(time.Minute * 20)
		for _, entry := range policy.Schedules.RecurringSchedules {
			if endtime.Day() != starttime.Day() {
				entry.StartTime = "00:01"
				entry.EndTime = "23:59"
				entry.StartDate = endtime.Format("2006-01-02")
			} else {
				entry.StartTime = starttime.Format("15:04")
				entry.EndTime = endtime.Format("15:04")
			}
		}
	}

	content, err := json.Marshal(policy)
	Expect(err).NotTo(HaveOccurred())
	return content
}

func setPolicySpecificDateTime(policyByte []byte, start time.Duration, end time.Duration) string {
	timeZone := "GMT"
	location, _ := time.LoadLocation(timeZone)
	timeNowInTimeZone := time.Now().In(location)
	dateTimeFormat := "2006-01-02T15:04"
	startTime := timeNowInTimeZone.Add(start).Format(dateTimeFormat)
	endTime := timeNowInTimeZone.Add(end).Format(dateTimeFormat)

	return fmt.Sprintf(string(policyByte), timeZone, startTime, endTime)
}

func getScalingHistories(apiURL url.URL, pathVariables []string, parameters map[string]string) (*http.Response, error) {
	var getScalingHistoriesURL string
	By("getScalingHistories")
	httpClientTmp := httpClientForPublicApi
	apiURL.Path = fmt.Sprintf("/v1/apps/%s/scaling_histories", pathVariables[0])
	if len(parameters) > 0 {
		parsedURL, err := url.Parse(apiURL.String())
		Expect(err).ToNot(HaveOccurred())

		params := url.Values{}
		for paramName, paramValue := range parameters {
			params.Add(paramName, paramValue)
		}
		parsedURL.RawQuery = params.Encode()

		getScalingHistoriesURL = parsedURL.String()
	} else {
		getScalingHistoriesURL = apiURL.String()
	}

	req, err := http.NewRequest("GET", getScalingHistoriesURL, strings.NewReader(""))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "bearer fake-token")
	return httpClientTmp.Do(req)
}

func getAppAggregatedMetrics(apiURL url.URL, pathVariables []string, parameters map[string]string) (*http.Response, error) {
	urlParams := ""
	By("getAppAggregatedMetrics")
	httpClientTmp := httpClientForPublicApi
	if len(parameters) > 0 {
		urlParams += "any=any"
		for paramName, paramValue := range parameters {
			urlParams += "&" + paramName + "=" + paramValue
		}
	}
	apiURL.Path = fmt.Sprintf("/v1/apps/%s/aggregated_metric_histories/%s?%s", pathVariables[0], pathVariables[1], urlParams)
	req, err := http.NewRequest("GET", apiURL.String(), strings.NewReader(""))

	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "bearer fake-token")
	return httpClientTmp.Do(req)
}

func readPolicyFromFile(filename string) []byte {
	content, err := os.ReadFile(filename)
	Expect(err).NotTo(HaveOccurred())
	return content
}

func clearDatabase() {
	_, err := dbHelper.Exec("DELETE FROM policy_json")
	Expect(err).NotTo(HaveOccurred())

	_, err = dbHelper.Exec("DELETE FROM binding")
	Expect(err).NotTo(HaveOccurred())

	_, err = dbHelper.Exec("DELETE FROM service_instance")
	Expect(err).NotTo(HaveOccurred())

	_, err = dbHelper.Exec("DELETE FROM app_scaling_recurring_schedule")
	Expect(err).NotTo(HaveOccurred())

	_, err = dbHelper.Exec("DELETE FROM app_scaling_specific_date_schedule")
	Expect(err).NotTo(HaveOccurred())

	_, err = dbHelper.Exec("DELETE FROM app_scaling_active_schedule")
	Expect(err).NotTo(HaveOccurred())

	_, err = dbHelper.Exec("DELETE FROM activeschedule")
	Expect(err).NotTo(HaveOccurred())

	_, err = dbHelper.Exec("DELETE FROM scalinghistory")
	Expect(err).NotTo(HaveOccurred())

	_, err = dbHelper.Exec("DELETE FROM app_metric")
	Expect(err).NotTo(HaveOccurred())
}

func insertPolicy(appId string, policyStr string, guid string) {
	query := dbHelper.Rebind("INSERT INTO policy_json(app_id, policy_json, guid) VALUES(?, ?, ?)")
	_, err := dbHelper.Exec(query, appId, policyStr, guid)
	Expect(err).NotTo(HaveOccurred())
}

func deletePolicy(appId string) {
	query := dbHelper.Rebind("DELETE FROM policy_json WHERE app_id=?")
	_, err := dbHelper.Exec(query, appId)
	Expect(err).NotTo(HaveOccurred())
}

func insertScalingHistory(history *models.AppScalingHistory) {
	query := dbHelper.Rebind("INSERT INTO scalinghistory" +
		"(appid, timestamp, scalingtype, status, oldinstances, newinstances, reason, message, error) " +
		" VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)")
	_, err := dbHelper.Exec(query, history.AppId, history.Timestamp, history.ScalingType, history.Status,
		history.OldInstances, history.NewInstances, history.Reason, history.Message, history.Error)

	Expect(err).NotTo(HaveOccurred())
}

func createScalingHistory(appId string, timestamp int64) models.AppScalingHistory {
	return models.AppScalingHistory{
		AppId:        appId,
		OldInstances: 2,
		NewInstances: 4,
		Reason:       "a reason",
		Message:      "a message",
		ScalingType:  models.ScalingTypeDynamic,
		Status:       models.ScalingStatusSucceeded,
		Error:        "",
		Timestamp:    timestamp,
	}
}

func createScalingHistoryError(appId string, timestamp int64) models.AppScalingHistory {
	return models.AppScalingHistory{
		AppId:        appId,
		OldInstances: -1,
		NewInstances: -1,
		Reason:       "a reason",
		Message:      "a message",
		ScalingType:  models.ScalingTypeDynamic,
		Status:       models.ScalingStatusFailed,
		Error:        "an error",
		Timestamp:    timestamp,
	}
}

func getScalingHistoryCount(appId string, oldInstanceCount int, newInstanceCount int) int {
	var count int
	query := dbHelper.Rebind("SELECT COUNT(*) FROM scalinghistory WHERE appid=? AND oldinstances=? AND newinstances=?")
	err := dbHelper.QueryRow(query, appId, oldInstanceCount, newInstanceCount).Scan(&count)
	Expect(err).NotTo(HaveOccurred())
	return count
}

func getScalingHistoryTotalCount(appId string) int {
	var count int
	query := dbHelper.Rebind("SELECT COUNT(*) FROM scalinghistory WHERE appid=?")
	err := dbHelper.QueryRow(query, appId).Scan(&count)
	Expect(err).NotTo(HaveOccurred())
	return count
}

func insertAppMetric(appMetrics *models.AppMetric) {
	query := dbHelper.Rebind("INSERT INTO app_metric" +
		"(app_id, metric_type, unit, value, timestamp) " +
		"VALUES(?, ?, ?, ?, ?)")
	_, err := dbHelper.Exec(query, appMetrics.AppId, appMetrics.MetricType, appMetrics.Unit, appMetrics.Value, appMetrics.Timestamp)
	Expect(err).NotTo(HaveOccurred())
}

func getAppMetricTotalCount(appId string) int {
	var count int
	query := dbHelper.Rebind("SELECT COUNT(*) FROM app_metric WHERE app_id=?")
	err := dbHelper.QueryRow(query, appId).Scan(&count)
	Expect(err).NotTo(HaveOccurred())
	return count
}

type GetResponse func(id string, url url.URL, httpClient *http.Client) (*http.Response, error)
type GetResponseWithParameters func(url url.URL, pathVariables []string, parameters map[string]string) (*http.Response, error)

func checkResponseContent(getResponse GetResponse, id string, expectHttpStatus int, expectResponseMap map[string]interface{}, url url.URL, httpClient *http.Client) {
	resp, err := getResponse(id, url, httpClient)
	defer func() { _ = resp.Body.Close() }()
	checkResponse(resp, err, expectHttpStatus, expectResponseMap)
}

func checkPublicAPIResponseContentWithParameters(getResponseWithParameters GetResponseWithParameters, apiURL url.URL, pathVariables []string, parameters map[string]string, expectHttpStatus int, expectResponseMap map[string]interface{}) {
	resp, err := getResponseWithParameters(apiURL, pathVariables, parameters)
	defer func() { _ = resp.Body.Close() }()
	checkResponse(resp, err, expectHttpStatus, expectResponseMap)
}

func checkResponse(resp *http.Response, err error, expectHttpStatus int, expectResponseMap map[string]interface{}) {
	Expect(err).WithOffset(2).NotTo(HaveOccurred())
	Expect(resp.StatusCode).WithOffset(2).To(Equal(expectHttpStatus))
	var actual map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&actual)
	Expect(err).WithOffset(2).NotTo(HaveOccurred())
	Expect(actual).WithOffset(2).To(Equal(expectResponseMap))
}

func checkResponseEmptyAndStatusCode(resp *http.Response, err error, expectedStatus int) {
	Expect(err).NotTo(HaveOccurred())
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())
	Expect(body).To(HaveLen(0))
	Expect(resp.StatusCode).To(Equal(expectedStatus))
}

func assertScheduleContents(schedulerURL url.URL, appId string, expectHttpStatus int, expectResponseMap map[string]int) {
	By("checking the schedule contents")
	resp, err := getSchedules(schedulerURL, appId)
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "Failed to get schedule:%s", err)
	ExpectWithOffset(1, resp.StatusCode).To(Equal(expectHttpStatus), "Unexpected HTTP status")
	defer func() { _ = resp.Body.Close() }()
	var actual map[string]interface{}

	err = json.NewDecoder(resp.Body).Decode(&actual)
	Expect(err).NotTo(HaveOccurred(), "Invalid JSON")

	var schedules = actual["schedules"].(map[string]interface{})
	var recurring = schedules["recurring_schedule"].([]interface{})
	var specificDate = schedules["specific_date"].([]interface{})
	ExpectWithOffset(1, len(specificDate)).To(Equal(expectResponseMap["specific_date"]), "Expected %d specific date schedules, but found %d: %#v\n", expectResponseMap["specific_date"], len(specificDate), specificDate)
	ExpectWithOffset(1, len(recurring)).To(Equal(expectResponseMap["recurring_schedule"]), "Expected %d recurring schedules, but found %d: %#v\n", expectResponseMap["recurring_schedule"], len(recurring), recurring)
}

func checkScheduleContents(schedulerURL url.URL, appId string, expectHttpStatus int, expectResponseMap map[string]int) bool {
	resp, err := getSchedules(schedulerURL, appId)
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "Get schedules failed with: %s", err)
	ExpectWithOffset(1, resp.StatusCode).To(Equal(expectHttpStatus), "Unexpected HTTP status")
	defer func() { _ = resp.Body.Close() }()
	var actual map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&actual)
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "Invalid JSON")
	var schedules = actual["schedules"].(map[string]interface{})
	var recurring = schedules["recurring_schedule"].([]interface{})
	var specificDate = schedules["specific_date"].([]interface{})
	return len(specificDate) == expectResponseMap["specific_date"] && len(recurring) == expectResponseMap["recurring_schedule"]
}

func startFakeCCNOAAUAA(instanceCount int) {
	fakeCCNOAAUAA = mocks.NewServer()
	fakeCCNOAAUAA.Add().
		GetApp(models.AppStatusStarted, http.StatusOK, "test_space_guid").
		GetAppProcesses(instanceCount).
		ScaleAppWebProcess().
		Roles(http.StatusOK, cf.Role{Type: cf.RoleSpaceDeveloper}).
		ServiceInstance("cc-free-plan-id").
		ServicePlan("autoscaler-free-plan-id").
		Info(fakeCCNOAAUAA.URL()).
		OauthToken(testUserToken).
		CheckToken(testUserScope).
		UserInfo(http.StatusOK, testUserId)
}
