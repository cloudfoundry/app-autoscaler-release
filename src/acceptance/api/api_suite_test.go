package api

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"acceptance/config"
	. "acceptance/helpers"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

const (
	HealthPath           = "/health"
	PolicyPath           = "/v1/apps/{appId}/policy"
	MetricPath           = "/v1/apps/{appId}/metric_histories/{metric_type}"
	AggregatedMetricPath = "/v1/apps/{appId}/aggregated_metric_histories/{metric_type}"
	HistoryPath          = "/v1/apps/{appId}/scaling_histories"
)

var (
	cfg                 *config.Config
	setup               *workflowhelpers.ReproducibleTestSuiteSetup
	appName             string
	appGUID             string
	instanceName        string
	healthURL           string
	policyURL           string
	metricURL           string
	aggregatedMetricURL string
	historyURL          string
	client              *http.Client
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)

	cfg = config.LoadConfig(t)
	componentName := "Public API Suite"
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
		if cfg.IsServiceOfferingEnabled() {
			EnableServiceAccess(cfg, setup.GetOrganizationName())
		}
	})

	appName = generator.PrefixedRandomName("autoscaler", "nodeapp")
	initialInstanceCount := 1
	countStr := strconv.Itoa(initialInstanceCount)
	createApp := cf.Cf("push", appName, "--no-start", "-i", countStr, "-b", cfg.NodejsBuildpackName, "-m", "128M", "-p", config.NODE_APP, "-d", cfg.AppsDomain).Wait(cfg.CfPushTimeoutDuration())
	Expect(createApp).To(Exit(0), "failed creating app")

	guid := cf.Cf("app", appName, "--guid").Wait(cfg.DefaultTimeoutDuration())
	Expect(guid).To(Exit(0))
	appGUID = strings.TrimSpace(string(guid.Out.Contents()))

	Expect(cf.Cf("start", appName).Wait(cfg.CfPushTimeoutDuration())).To(Exit(0))
	WaitForNInstancesRunning(appGUID, initialInstanceCount, cfg.DefaultTimeoutDuration())

	if cfg.IsServiceOfferingEnabled() {
		serviceExists := cf.Cf("marketplace", "-s", cfg.ServiceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(serviceExists).To(Exit(0), fmt.Sprintf("Service offering, %s, does not exist", cfg.ServiceName))

		instanceName = generator.PrefixedRandomName("autoscaler", "service")
		createService := cf.Cf("create-service", cfg.ServiceName, cfg.ServicePlan, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(createService).To(Exit(0), "failed creating service")

		bindService := cf.Cf("bind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(bindService).To(Exit(0), "failed binding service to app ")
	}

	client = &http.Client{
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

	healthURL = fmt.Sprintf("%s%s", cfg.ASApiEndpoint, HealthPath)
	policyURL = fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(PolicyPath, "{appId}", appGUID, -1))
	metricURL = strings.Replace(MetricPath, "{metric_type}", "memoryused", -1)
	metricURL = fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(metricURL, "{appId}", appGUID, -1))
	aggregatedMetricURL = strings.Replace(AggregatedMetricPath, "{metric_type}", "memoryused", -1)
	aggregatedMetricURL = fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(aggregatedMetricURL, "{appId}", appGUID, -1))
	historyURL = fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(HistoryPath, "{appId}", appGUID, -1))

})

var _ = AfterSuite(func() {

	if cfg.IsServiceOfferingEnabled() {
		unbindService := cf.Cf("unbind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(unbindService).To(Exit(0), "failed unbinding service from app")

		deleteService := cf.Cf("delete-service", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
		Expect(deleteService).To(Exit(0))
	}

	Expect(cf.Cf("delete", appName, "-f", "-r").Wait(cfg.DefaultTimeoutDuration())).To(Exit(0))
	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		if cfg.IsServiceOfferingEnabled() {
			DisableServiceAccess(cfg, setup.GetOrganizationName())
		}
	})
	setup.Teardown()
})

func DoAPIRequest(req *http.Request) (*http.Response, error) {
	resp, err := client.Do(req)
	return resp, err
}
