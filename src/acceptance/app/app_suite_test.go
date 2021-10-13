package app_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
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

const (
	CustomMetricPath    = "/v1/apps/{appId}/credential"
	CustomMetricCredEnv = "AUTO_SCALER_CUSTOM_METRIC_ENV"
)

var (
	cfg      *config.Config
	setup    *workflowhelpers.ReproducibleTestSuiteSetup
	interval int
	client   *http.Client

	instanceName         string
	initialInstanceCount int
)

type CFResourceObject struct {
	Resources []struct {
		GUID      string `json:"guid"`
		CreatedAt string `json:"created_at"`
		Name      string `json:"name"`
		Username  string `json:"username"`
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
	componentName := "Application Scale Suite"

	if cfg.GetArtifactsDirectory() != "" {
		helpers.EnableCFTrace(cfg, componentName)
	}

	RunSpecs(t, componentName)
}

var _ = BeforeSuite(func() {

	setup = workflowhelpers.NewTestSuiteSetup(cfg)

	Cleanup(cfg, setup)

	setup.Setup()

	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		if cfg.IsServiceOfferingEnabled() && cfg.ShouldEnableServiceAccess() {
			EnableServiceAccess(cfg, setup.GetOrganizationName())
		}
	})

	if cfg.IsServiceOfferingEnabled() {
		CheckServiceExists(cfg)
	}

	interval = cfg.AggregateInterval

	client = GetHTTPClient(cfg)

})

var _ = AfterSuite(func() {
	if os.Getenv("SKIP_TEARDOWN") == "true" {
		fmt.Println("Skipping Teardown...")
	} else {
		workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
			if cfg.IsServiceOfferingEnabled() && cfg.ShouldEnableServiceAccess() {
				DisableServiceAccess(cfg, setup.GetOrganizationName())
			}
		})
		setup.Teardown()
	}
})

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

func doAPIRequest(req *http.Request) (*http.Response, error) {
	return client.Do(req)
}

func DeletePolicyWithAPI(appGUID string) {
	oauthToken := OauthToken(cfg)
	policyURL := fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(PolicyPath, "{appId}", appGUID, -1))
	req, err := http.NewRequest("DELETE", policyURL, nil)
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)

	resp, err := doAPIRequest(req)
	Expect(err).ShouldNot(HaveOccurred())
	defer resp.Body.Close()
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
}

func DeletePolicy(appName, appGUID string) {
	if cfg.IsServiceOfferingEnabled() {
		unbindService := cf.Cf("unbind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(unbindService).To(Exit(0), "failed unbinding service from app")
		deleteService := cf.Cf("delete-service", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
		Expect(deleteService).To(Exit(0))
	} else {
		DeletePolicyWithAPI(appGUID)
	}
}

func CreateCustomMetricCred(appName, appGUID string) {
	oauthToken := OauthToken(cfg)
	customMetricURL := fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(CustomMetricPath, "{appId}", appGUID, -1))
	req, err := http.NewRequest("PUT", customMetricURL, nil)
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)

	resp, err := doAPIRequest(req)
	Expect(err).ShouldNot(HaveOccurred())
	defer resp.Body.Close()
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())
	setEnv := cf.Cf("set-env", appName, CustomMetricCredEnv, string(bodyBytes)).Wait(cfg.DefaultTimeoutDuration())
	Expect(setEnv).To(Exit(0), "failed set custom metric credential env")
}

func DeleteCustomMetricCred(appGUID string) {
	oauthToken := OauthToken(cfg)
	customMetricURL := fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(CustomMetricPath, "{appId}", appGUID, -1))
	req, err := http.NewRequest("DELETE", customMetricURL, nil)
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)

	resp, err := doAPIRequest(req)
	Expect(err).ShouldNot(HaveOccurred())
	defer resp.Body.Close()
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
}
