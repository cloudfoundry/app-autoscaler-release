package api_test

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"acceptance/config"
	. "acceptance/helpers"

	"github.com/KevinJCross/cf-test-helpers/v2/cf"
	"github.com/KevinJCross/cf-test-helpers/v2/generator"
	"github.com/KevinJCross/cf-test-helpers/v2/helpers"
	"github.com/KevinJCross/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

const (
	HealthPath           = "/health"
	MetricPath           = "/v1/apps/{appId}/metric_histories/{metric_type}"
	AggregatedMetricPath = "/v1/apps/{appId}/aggregated_metric_histories/{metric_type}"
	HistoryPath          = "/v1/apps/{appId}/scaling_histories"
)

var (
	cfg                 *config.Config
	setup               *workflowhelpers.ReproducibleTestSuiteSetup
	otherSetup          *workflowhelpers.ReproducibleTestSuiteSetup
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

	if cfg.GetArtifactsDirectory() != "" {
		helpers.EnableCFTrace(cfg, componentName)
	}

	RunSpecs(t, componentName)
}

var _ = BeforeSuite(func() {

	otherConfig := *cfg
	otherConfig.NamePrefix = otherConfig.NamePrefix + "_other"

	By("Setup test environment")
	setup = workflowhelpers.NewTestSuiteSetup(cfg)
	otherSetup = workflowhelpers.NewTestSuiteSetup(&otherConfig)

	//Cleanup(cfg, setup)
	//Cleanup(&otherConfig, otherSetup)

	otherSetup.Setup()
	setup.Setup()

	By("Enable Service Access")
	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		if cfg.IsServiceOfferingEnabled() && cfg.ShouldEnableServiceAccess() {
			EnableServiceAccess(cfg, setup.GetOrganizationName())
		}
	})

	initialInstanceCount := 1
	appName = CreateTestApp(cfg, appName, initialInstanceCount)
	appGUID = GetAppGuid(cfg, appName)
	StartApp(appName, cfg.CfPushTimeoutDuration())
	WaitForNInstancesRunning(appGUID, initialInstanceCount, cfg.DefaultTimeoutDuration())

	By("Creating test service")
	if cfg.IsServiceOfferingEnabled() {
		CheckServiceExists(cfg, setup.TestSpace.SpaceName(), cfg.ServiceName)

		instanceName = generator.PrefixedRandomName(cfg.Prefix, cfg.InstancePrefix)
		AbortOnCommandFailuref(
			cf.Cf("create-service", cfg.ServiceName, cfg.ServicePlan, instanceName, "-b", cfg.ServiceBroker).
				Wait(cfg.DefaultTimeoutDuration()),
			"failed creating service %s", cfg.ServiceName)

		AbortOnCommandFailuref(
			cf.Cf("bind-service", appName, instanceName).
				Wait(cfg.DefaultTimeoutDuration()),
			"failed binding service %s to app %s", instanceName, appName)
	}

	// #nosec G402
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
	if os.Getenv("SKIP_TEARDOWN") == "true" {
		fmt.Println("Skipping Teardown...")
	} else {
		if cfg.IsServiceOfferingEnabled() {
			if appName != "" && instanceName != "" {
				unbindService := cf.Cf("unbind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
				if unbindService.ExitCode() != 0 {
					purgeService := cf.Cf("purge-service-instance", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
					Expect(purgeService).To(Exit(0), fmt.Sprintf("failed to purge service instance %s", instanceName))
				}
			}

			if instanceName != "" {
				deleteService := cf.Cf("delete-service", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
				if deleteService.ExitCode() != 0 {
					purgeService := cf.Cf("purge-service-instance", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
					Expect(purgeService).To(Exit(0), fmt.Sprintf("failed to purge service instance %s", instanceName))
				}
			}
		}

		if appName != "" {
			deleteApp := cf.Cf("delete", appName, "-f", "-r").Wait(cfg.DefaultTimeoutDuration())
			Expect(deleteApp).To(Exit(0), fmt.Sprintf("unable to delete app %s", appName))
		}

		workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
			if cfg.IsServiceOfferingEnabled() && cfg.ShouldEnableServiceAccess() {
				DisableServiceAccess(cfg, setup.GetOrganizationName())
			}
		})

		otherSetup.Teardown()
		setup.Teardown()
	}
})

func DoAPIRequest(req *http.Request) (*http.Response, error) {
	return client.Do(req)
}
