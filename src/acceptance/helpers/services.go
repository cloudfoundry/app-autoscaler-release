package helpers

import (
	"acceptance/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/KevinJCross/cf-test-helpers/v2/cf"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type cfResourceObject struct {
	Resources []struct {
		GUID      string `json:"guid"`
		CreatedAt string `json:"created_at"`
		Name      string `json:"name"`
		Username  string `json:"username"`
	} `json:"resources"`
}

func GetServices(cfg *config.Config, orgGuid, spaceGuid string, prefix string) []string {
	var services cfResourceObject
	rawServices := cf.Cf("curl", "/v3/service_instances?space_guids="+spaceGuid+"&organization_guids="+orgGuid).Wait(cfg.DefaultTimeoutDuration())
	Expect(rawServices).To(Exit(0), "unable to get services")
	err := json.Unmarshal(rawServices.Out.Contents(), &services)
	Expect(err).ShouldNot(HaveOccurred())

	return filterByPrefix(prefix, getNames(services))
}

func DeleteServices(cfg *config.Config, services []string) {
	for _, service := range services {
		deleteService := cf.Cf("delete-service", service, "-f").Wait(cfg.DefaultTimeoutDuration())
		if deleteService.ExitCode() != 0 {
			fmt.Printf("unable to delete the service %s, attempting to purge...\n", service)
			purgeService := cf.Cf("purge-service-instance", service, "-f").Wait(cfg.DefaultTimeoutDuration())
			Expect(purgeService).To(Exit(0), fmt.Sprintf("unable to delete service %s", service))
		}
	}
}

const (
	CustomMetricPath    = "/v1/apps/{appId}/credential"
	CustomMetricCredEnv = "AUTO_SCALER_CUSTOM_METRIC_ENV" // #nosec G101
)

func CreateCustomMetricCred(cfg *config.Config, appName, appGUID string) {
	oauthToken := OauthToken(cfg)
	customMetricURL := fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(CustomMetricPath, "{appId}", appGUID, -1))
	req, err := http.NewRequest("PUT", customMetricURL, nil)
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)

	resp, err := GetHTTPClient(cfg).Do(req)
	Expect(err).ShouldNot(HaveOccurred())
	defer func() { _ = resp.Body.Close() }()
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())
	setEnv := cf.Cf("set-env", appName, CustomMetricCredEnv, string(bodyBytes)).Wait(cfg.DefaultTimeoutDuration())
	Expect(setEnv).To(Exit(0), "failed set custom metric credential env")
}

func DeleteCustomMetricCred(cfg *config.Config, appGUID string) {
	oauthToken := OauthToken(cfg)
	customMetricURL := fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(CustomMetricPath, "{appId}", appGUID, -1))
	req, err := http.NewRequest("DELETE", customMetricURL, nil)
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)

	resp, err := GetHTTPClient(cfg).Do(req)
	Expect(err).ShouldNot(HaveOccurred())
	defer func() { _ = resp.Body.Close() }()
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
}
