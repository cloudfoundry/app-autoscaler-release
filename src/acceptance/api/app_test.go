package api_test

import (
	"acceptance/config"
	. "acceptance/helpers"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

const (
	HealthPath           = "/health"
	MetricPath           = "/v1/apps/{appId}/metric_histories/{metric_type}"
	AggregatedMetricPath = "/v1/apps/{appId}/aggregated_metric_histories/{metric_type}"
	HistoryPath          = "/v1/apps/{appId}/scaling_histories"
)

type AppInstanceMetric struct {
	AppId         string `json:"app_id"`
	InstanceIndex uint32 `json:"instance_index"`
	CollectedAt   int64  `json:"collected_at"`
	Name          string `json:"name"`
	Unit          string `json:"unit"`
	Value         string `json:"value"`
	Timestamp     int64  `json:"timestamp"`
}

type AppMetric struct {
	AppId      string `json:"app_id"`
	MetricType string `json:"name"`
	Value      string `json:"value"`
	Unit       string `json:"unit"`
	Timestamp  int64  `json:"timestamp"`
}

type AppScalingHistory struct {
	AppId        string `json:"app_id"`
	Timestamp    int64  `json:"timestamp"`
	ScalingType  int    `json:"scaling_type"`
	Status       int    `json:"status"`
	OldInstances int    `json:"old_instances"`
	NewInstances int    `json:"new_instances"`
	Reason       string `json:"reason"`
	Message      string `json:"message"`
	Error        string `json:"error"`
}

type MetricsResults struct {
	TotalResults uint32               `json:"total_results"`
	TotalPages   uint16               `json:"total_pages"`
	Page         uint16               `json:"page"`
	Metrics      []*AppInstanceMetric `json:"resources"`
}

type AggregatedMetricsResults struct {
	TotalResults uint32       `json:"total_results"`
	TotalPages   uint16       `json:"total_pages"`
	Page         uint16       `json:"page"`
	Metrics      []*AppMetric `json:"resources"`
}

type HistoryResults struct {
	TotalResults uint32               `json:"total_results"`
	TotalPages   uint16               `json:"total_pages"`
	Page         uint16               `json:"page"`
	Histories    []*AppScalingHistory `json:"resources"`
}

type App struct {
	isCreated            bool
	initialInstanceCount int
	name                 string
	GUID                 string
	policyURL            string
	metricURL            string
	aggregatedMetricURL  string
	historyURL           string
	client               *http.Client
}

func New(cfg *config.Config) *App {
	return &App{
		name: generator.PrefixedRandomName(cfg.Prefix, cfg.AppPrefix),
	}
}

func (a *App) Create(instances int) {
	a.initialInstanceCount = instances
	countStr := strconv.Itoa(a.initialInstanceCount)
	createApp := cf.Cf("push", a.name, "--no-start", "--no-route", "-i", countStr, "-b", cfg.NodejsBuildpackName, "-m", "128M", "-p", config.NODE_APP).Wait(cfg.CfPushTimeoutDuration())
	Expect(createApp).To(Exit(0), "failed creating app")

	mapRouteToApp := cf.Cf("map-route", a.name, cfg.AppsDomain, "--hostname", a.name).Wait(cfg.DefaultTimeoutDuration())
	Expect(mapRouteToApp).To(Exit(0), "failed to map route to app")

	guid := cf.Cf("app", a.name, "--guid").Wait(cfg.DefaultTimeoutDuration())
	Expect(guid).To(Exit(0))
	a.GUID = strings.TrimSpace(string(guid.Out.Contents()))
	a.isCreated = true

	// #nosec G402
	a.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
			DisableCompression:  true,
			DisableKeepAlives:   true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.SkipSSLValidation,
			},
		},
		Timeout: 30 * time.Second,
	}

	a.policyURL = fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(PolicyPath, "{appId}", a.GUID, -1))
	a.metricURL = strings.Replace(MetricPath, "{metric_type}", "memoryused", -1)
	a.metricURL = fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(a.metricURL, "{appId}", a.GUID, -1))
	a.aggregatedMetricURL = strings.Replace(AggregatedMetricPath, "{metric_type}", "memoryused", -1)
	a.aggregatedMetricURL = fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(a.aggregatedMetricURL, "{appId}", a.GUID, -1))
	a.historyURL = fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(HistoryPath, "{appId}", a.GUID, -1))
}

func (a *App) Start() {
	Expect(cf.Cf("start", a.name).Wait(cfg.CfPushTimeoutDuration())).To(Exit(0))
	WaitForNInstancesRunning(a.GUID, app.initialInstanceCount, cfg.DefaultTimeoutDuration())
}

func (a *App) Delete() {
	if a != nil && a.isCreated {
		deleteApp := cf.Cf("delete", a.name, "-f", "-r").Wait(cfg.DefaultTimeoutDuration())
		Expect(deleteApp).To(Exit(0), fmt.Sprintf("unable to delete app %s", app.name))
		a.isCreated = false
	}
}

func (a *App) History() *HistoryResults {
	raw, status := a.get(a.historyURL)
	Expect(status).To(Equal(200))

	var histories *HistoryResults
	err := json.Unmarshal(raw, &histories)
	Expect(err).ShouldNot(HaveOccurred())
	return histories
}

func (a *App) AggregatedMetrics() *AggregatedMetricsResults {
	raw, status := a.get(a.aggregatedMetricURL)
	Expect(status).To(Equal(200))
	var metrics *AggregatedMetricsResults
	err := json.Unmarshal(raw, &metrics)
	Expect(err).ShouldNot(HaveOccurred())
	return metrics
}

func (a *App) Metrics() *MetricsResults {
	raw, status := a.get(a.metricURL)
	Expect(status).To(Equal(200))

	var metrics *MetricsResults
	err := json.Unmarshal(raw, &metrics)
	Expect(err).ShouldNot(HaveOccurred())
	return metrics
}

func (a *App) CreatePolicy(policy string) ([]byte, int) {
	return a.put(a.policyURL, policy)
}

func (a *App) put(url string, body string) ([]byte, int) {
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer([]byte(body)))
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	Expect(err).ShouldNot(HaveOccurred())

	defer func() { _ = resp.Body.Close() }()

	raw, err := ioutil.ReadAll(resp.Body)
	Expect(err).ShouldNot(HaveOccurred())
	return raw, resp.StatusCode
}

func (a *App) DeletePolicy() ([]byte, int) {
	return a.deleteReq(a.policyURL)
}

func (a *App) deleteReq(url string) ([]byte, int) {
	//delete policy here to make sure the condition "no policy defined"
	req, err := http.NewRequest("DELETE", url, nil)
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	Expect(err).ShouldNot(HaveOccurred())
	defer func() { _ = resp.Body.Close() }()
	response, err := ioutil.ReadAll(resp.Body)
	Expect(err).ShouldNot(HaveOccurred())
	return response, resp.StatusCode
}

func (a *App) GetPolicy() ([]byte, int) {
	return a.get(a.policyURL)
}

func (a *App) get(url string) ([]byte, int) {
	req, err := http.NewRequest("GET", url, nil)
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	Expect(err).ShouldNot(HaveOccurred())

	defer func() { _ = resp.Body.Close() }()

	policy, err := ioutil.ReadAll(resp.Body)
	Expect(err).ShouldNot(HaveOccurred())
	return policy, resp.StatusCode
}
