package api_test

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
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

type CFResourceObject struct {
	Resources []struct {
		GUID      string `json:"guid"`
		CreatedAt string `json:"created_at"`
		Name      string `json:"name"`
		Username  string `json:"username"`
	} `json:"resources"`
}

type CFUsers struct {
	Resources []struct {
		Entity struct {
			Username string `json:"username"`
		}
		Metadata struct {
			GUID      string `json:"guid"`
			CreatedAt string `json:"created_at"`
		}
	} `json:"resources"`
}

type CFOrgs struct {
	Resources []struct {
		Name      string `json:"name"`
		GUID      string `json:"guid"`
		CreatedAt string `json:"created_at"`
	} `json:"resources"`
}

type CFSpaces struct {
	Resources []struct {
		Entity struct {
			Name string `json:"name"`
		}
		Metadata struct {
			GUID      string `json:"guid"`
			CreatedAt string `json:"created_at"`
		}
	} `json:"resources"`
}

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

	fmt.Println("Clearing down existing test orgs/spaces...")
	setup = workflowhelpers.NewTestSuiteSetup(cfg)

	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		orgs := getTestOrgs()

		for _, org := range orgs {
			orgName, orgGuid, spaceName, spaceGuid := getOrgSpaceNamesAndGuids(org)
			if spaceName != "" {
				target := cf.Cf("target", "-o", orgName, "-s", spaceName).Wait(cfg.DefaultTimeoutDuration())
				Expect(target).To(Exit(0), fmt.Sprintf("failed to target %s and %s", orgName, spaceName))

				apps := getApps(orgGuid, spaceGuid, "autoscaler-")
				deleteApps(apps, 0)

				services := getServices(orgGuid, spaceGuid, "autoscaler-")
				deleteServices(services)
			}

			deleteOrg(org)
		}
	})

	fmt.Println("Clearing down existing test orgs/spaces... Complete")

	setup.Setup()

	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		if cfg.IsServiceOfferingEnabled() && cfg.ShouldEnableServiceAccess() {
			EnableServiceAccess(cfg, setup.GetOrganizationName())
		}
	})

	appName = generator.PrefixedRandomName("autoscaler", "nodeapp")
	initialInstanceCount := 1
	countStr := strconv.Itoa(initialInstanceCount)
	createApp := cf.Cf("push", appName, "--no-start", "--no-route", "-i", countStr, "-b", cfg.NodejsBuildpackName, "-m", "128M", "-p", config.NODE_APP).Wait(cfg.CfPushTimeoutDuration())
	Expect(createApp).To(Exit(0), "failed creating app")

	mapRouteToApp := cf.Cf("map-route", appName, cfg.AppsDomain, "--hostname", appName).Wait(cfg.DefaultTimeoutDuration())
	Expect(mapRouteToApp).To(Exit(0), "failed to map route to app")

	guid := cf.Cf("app", appName, "--guid").Wait(cfg.DefaultTimeoutDuration())
	Expect(guid).To(Exit(0))
	appGUID = strings.TrimSpace(string(guid.Out.Contents()))

	Expect(cf.Cf("start", appName).Wait(cfg.CfPushTimeoutDuration())).To(Exit(0))
	WaitForNInstancesRunning(appGUID, initialInstanceCount, cfg.DefaultTimeoutDuration())

	if cfg.IsServiceOfferingEnabled() {
		CheckServiceExists(cfg)

		instanceName = generator.PrefixedRandomName("autoscaler", "service")
		createService := cf.Cf("create-service", cfg.ServiceName, cfg.ServicePlan, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(createService).To(Exit(0), fmt.Sprintf("failed creating service %s", instanceName))

		bindService := cf.Cf("bind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(bindService).To(Exit(0), fmt.Sprintf("failed binding service %s to app %s", instanceName, appName))
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

		setup.Teardown()
	}
})

func DoAPIRequest(req *http.Request) (*http.Response, error) {
	return client.Do(req)
}

func getTestOrgs() []string {
	rawOrgs := cf.Cf("curl", "/v3/organizations").Wait(cfg.DefaultTimeoutDuration())
	Expect(rawOrgs).To(Exit(0), "unable to get orgs")

	var orgs CFOrgs
	err := json.Unmarshal(rawOrgs.Out.Contents(), &orgs)
	Expect(err).ShouldNot(HaveOccurred())

	var orgNames []string
	for _, org := range orgs.Resources {
		if strings.HasPrefix(org.Name, cfg.NamePrefix) {
			orgNames = append(orgNames, org.Name)
		}
	}

	return orgNames
}

func getOrgSpaceNamesAndGuids(org string) (string, string, string, string) {
	orgGuidByte := cf.Cf("org", org, "--guid").Wait(cfg.DefaultTimeoutDuration())
	orgGuid := strings.TrimSuffix(string(orgGuidByte.Out.Contents()), "\n")

	rawSpaces := cf.Cf("curl", fmt.Sprintf("/v2/organizations/%s/spaces", orgGuid)).Wait(cfg.DefaultTimeoutDuration())
	Expect(rawSpaces).To(Exit(0), "unable to get spaces")
	var spaces CFSpaces
	err := json.Unmarshal(rawSpaces.Out.Contents(), &spaces)
	Expect(err).ShouldNot(HaveOccurred())
	if len(spaces.Resources) == 0 {
		return org, orgGuid, "", ""
	}

	return org, orgGuid, spaces.Resources[0].Entity.Name, spaces.Resources[0].Metadata.GUID
}

func getServices(orgGuid, spaceGuid string, prefix string) []string {
	var services CFResourceObject
	rawServices := cf.Cf("curl", "/v3/service_instances?space_guids="+spaceGuid+"&organization_guids="+orgGuid).Wait(cfg.DefaultTimeoutDuration())
	Expect(rawServices).To(Exit(0), "unable to get services")
	err := json.Unmarshal(rawServices.Out.Contents(), &services)
	Expect(err).ShouldNot(HaveOccurred())

	var names []string
	for _, service := range services.Resources {
		if strings.HasPrefix(service.Name, prefix) {
			names = append(names, service.Name)
		}
	}

	return names
}

func getApps(orgGuid, spaceGuid string, prefix string) []string {
	var apps CFResourceObject
	rawApps := cf.Cf("curl", "/v3/apps?space_guids="+spaceGuid+"&organization_guids="+orgGuid).Wait(cfg.DefaultTimeoutDuration())
	Expect(rawApps).To(Exit(0), "unable to get apps")
	err := json.Unmarshal(rawApps.Out.Contents(), &apps)
	Expect(err).ShouldNot(HaveOccurred())

	var names []string
	for _, app := range apps.Resources {
		if strings.HasPrefix(app.Name, prefix) {
			names = append(names, app.Name)
		}
	}

	return names
}

func deleteServices(services []string) {
	for _, service := range services {
		deleteService := cf.Cf("delete-service", service, "-f").Wait(3 * cfg.DefaultTimeoutDuration())
		if deleteService.ExitCode() != 0 {
			fmt.Printf("unable to delete the service %s, attempting to purge...\n", service)
			purgeService := cf.Cf("purge-service-instance", service, "-f").Wait(cfg.DefaultTimeoutDuration())
			Expect(purgeService).To(Exit(0), fmt.Sprintf("unable to delete service %s", service))
		}
	}
}

func deleteOrg(org string) {
	deleteOrg := cf.Cf("delete-org", org, "-f").Wait(3 * cfg.DefaultTimeoutDuration())
	Expect(deleteOrg).To(Exit(0), fmt.Sprintf("unable to delete org %s", org))
}

func deleteApps(apps []string, threshold int) {
	for _, app := range apps {
		deleteApp := cf.Cf("delete", app, "-f").Wait(3 * cfg.DefaultTimeoutDuration())
		if deleteApp.ExitCode() != 0 {
			fmt.Printf("unable to delete the app %s, attempting to purge...\n", app)
			//purgeService := cf.Cf("purge-service-instance", service, "-f").Wait(cfg.DefaultTimeoutDuration())
			//Expect(purgeService).To(Exit(0), fmt.Sprintf("unable to delete service %s", service))
		}
		//Expect(deleteApp).To(Exit(0), fmt.Sprintf("unable to delete app %s", app))
	}
}
