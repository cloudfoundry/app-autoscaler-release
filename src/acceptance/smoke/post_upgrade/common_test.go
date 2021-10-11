package post_upgrade_test

import (
	"acceptance/helpers"
	"fmt"
	"net/http"
	"strings"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

const (
	CustomMetricPath = "/v1/apps/{appId}/credential"
)

func DeletePolicy(appName, appGUID, instanceName string) {
	if cfg.IsServiceOfferingEnabled() {
		unbindService := cf.Cf("unbind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(unbindService).To(Exit(0), "failed unbinding service from app")
		deleteService := cf.Cf("delete-service", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
		Expect(deleteService).To(Exit(0))
	} else {
		DeletePolicyWithAPI(appGUID)
	}
}

func DeletePolicyWithAPI(appGUID string) {
	oauthToken := helpers.OauthToken(cfg)
	policyURL := fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(helpers.PolicyPath, "{appId}", appGUID, -1))
	req, err := http.NewRequest("DELETE", policyURL, nil)
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)

	resp, err := helpers.GetHTTPClient(cfg).Do(req)
	Expect(err).ShouldNot(HaveOccurred())
	defer func() { _ = resp.Body.Close() }()
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
}

func DeleteCustomMetricCred(appGUID string) {
	oauthToken := helpers.OauthToken(cfg)
	customMetricURL := fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(CustomMetricPath, "{appId}", appGUID, -1))
	req, err := http.NewRequest("DELETE", customMetricURL, nil)
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)

	resp, err := helpers.GetHTTPClient(cfg).Do(req)
	Expect(err).ShouldNot(HaveOccurred())
	defer func() { _ = resp.Body.Close() }()
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
}

func GetAppInfo(org, space, appType string) (fullAppName string, appGuid string) {
	apps := helpers.GetApps(cfg, org, space, "autoscaler-")
	for _, app := range apps {
		if strings.Contains(app, appType) {
			return app, helpers.GetAppGuid(cfg, app)
		}
	}
	return "", ""
}
