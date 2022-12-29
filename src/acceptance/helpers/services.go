package helpers

import (
	"acceptance/config"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/KevinJCross/cf-test-helpers/v2/workflowhelpers"

	"github.com/KevinJCross/cf-test-helpers/v2/cf"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type cfResourceObject struct {
	Pagination struct {
		TotalPages int `json:"total_pages"`
		Next       struct {
			Href string `json:"href"`
		} `json:"next"`
	} `json:"pagination"`
	Resources []cfResource `json:"resources"`
}

type cfResource struct {
	GUID      string `json:"guid"`
	CreatedAt string `json:"created_at"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	State     string `json:"state"`
}

const (
	CustomMetricPath    = "/v1/apps/{appId}/credential"
	CustomMetricCredEnv = "AUTO_SCALER_CUSTOM_METRIC_ENV" // #nosec G101
)

func CreateCustomMetricCred(cfg *config.Config, appName, appGUID string) {
	if !cfg.IsServiceOfferingEnabled() {
		oauthToken := OauthToken(cfg)
		customMetricURL := fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(CustomMetricPath, "{appId}", appGUID, -1))
		req, err := http.NewRequest("PUT", customMetricURL, nil)
		Expect(err).ShouldNot(HaveOccurred())
		req.Header.Add("Authorization", oauthToken)

		//TODO ... this wont scale to 1000 apps at once
		resp, err := GetHTTPClient(cfg).Do(req)
		Expect(err).ShouldNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		bodyBytes, err := io.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())
		setEnv := cf.Cf("set-env", appName, CustomMetricCredEnv, string(bodyBytes)).Wait(cfg.DefaultTimeoutDuration())
		Expect(setEnv).To(Exit(0), "failed set custom metric credential env")
	}
}

func DeleteCustomMetricCred(cfg *config.Config, appGUID string) {
	if !cfg.IsServiceOfferingEnabled() {
		oauthToken := OauthToken(cfg)
		customMetricURL := fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(CustomMetricPath, "{appId}", appGUID, -1))
		req, err := http.NewRequest("DELETE", customMetricURL, nil)
		Expect(err).ShouldNot(HaveOccurred())
		req.Header.Add("Authorization", oauthToken)

		resp, err := GetHTTPClient(cfg).Do(req)
		Expect(err).ShouldNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
	}
}

func DeleteService(cfg *config.Config, setup *workflowhelpers.ReproducibleTestSuiteSetup, instanceName, appName string) {
	if cfg.IsServiceOfferingEnabled() {
		if appName != "" && instanceName != "" {
			UnbindService(cfg, instanceName, appName)
		}

		if instanceName != "" {
			deleteService := cf.Cf("delete-service", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
			if deleteService.ExitCode() != 0 {
				PurgeService(cfg, setup, instanceName)
			}
		}
	}
}

func UnbindService(cfg *config.Config, instanceName string, appName string) {
	unbindService := cf.Cf("unbind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
	if unbindService.ExitCode() != 0 {
		PurgeService(cfg, nil, instanceName)
	}
}

func PurgeService(cfg *config.Config, setup *workflowhelpers.ReproducibleTestSuiteSetup, instanceName string) {
	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		purgeService := cf.Cf("purge-service-instance", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
		Expect(purgeService).To(Exit(0), fmt.Sprintf("failed to purge service instance %s", instanceName))
	})
}
