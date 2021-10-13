package pre_upgrade_test

import (
	"acceptance/helpers"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

const (
	CustomMetricPath    = "/v1/apps/{appId}/credential"
	CustomMetricCredEnv = "AUTO_SCALER_CUSTOM_METRIC_ENV"
)

func CreateCustomMetricCred(appName, appGUID string) {
	oauthToken := helpers.OauthToken(cfg)
	customMetricURL := fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(CustomMetricPath, "{appId}", appGUID, -1))
	req, err := http.NewRequest("PUT", customMetricURL, nil)
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)

	resp, err := helpers.GetHTTPClient(cfg).Do(req)
	Expect(err).ShouldNot(HaveOccurred())
	defer func() { _ = resp.Body.Close() }()
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())
	setEnv := cf.Cf("set-env", appName, CustomMetricCredEnv, string(bodyBytes)).Wait(cfg.DefaultTimeoutDuration())
	Expect(setEnv).To(Exit(0), "failed set custom metric credential env")
}

func CreateApp(appType string, initialInstanceCount int) string {
	appName := generator.PrefixedRandomName("autoscaler", appType)
	countStr := strconv.Itoa(initialInstanceCount)
	createApp := cf.Cf("push", appName, "--no-start", "--no-route", "-i", countStr, "-b", cfg.NodejsBuildpackName, "-m", "128M", "-p", "../../assets/app/nodeApp").Wait(cfg.CfPushTimeoutDuration())
	Expect(createApp).To(Exit(0), "failed creating app")

	mapRouteToApp := cf.Cf("map-route", appName, cfg.AppsDomain, "--hostname", appName).Wait(cfg.DefaultTimeoutDuration())
	Expect(mapRouteToApp).To(Exit(0), "failed to map route to app")
	return appName
}

func StartApp(appName string) bool {
	return Expect(cf.Cf("start", appName).Wait(cfg.CfPushTimeoutDuration())).To(Exit(0))
}
