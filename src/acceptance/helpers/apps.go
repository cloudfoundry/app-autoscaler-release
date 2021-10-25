package helpers

import (
	"acceptance/config"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	cfh "github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func GetApps(cfg *config.Config, orgGuid, spaceGuid string, prefix string) []string {
	var apps cfResourceObject
	rawApps := cf.Cf("curl", "/v3/apps?space_guids="+spaceGuid+"&organization_guids="+orgGuid).Wait(cfg.DefaultTimeoutDuration())
	Expect(rawApps).To(Exit(0), "unable to get apps")
	err := json.Unmarshal(rawApps.Out.Contents(), &apps)
	Expect(err).ShouldNot(HaveOccurred())

	return filterByPrefix(prefix, getNames(apps))
}

func DeleteApps(cfg *config.Config, apps []string, threshold int) {
	for _, app := range apps {
		deleteApp := cf.Cf("delete", app, "-f").Wait(cfg.DefaultTimeoutDuration())
		Expect(deleteApp).To(Exit(0), fmt.Sprintf("unable to delete app %s", app))
	}
}

func SendMetric(cfg *config.Config, appName string, metric int) {
	Eventually(func() string {
		response := cfh.CurlApp(cfg, appName, fmt.Sprintf("/custom-metrics/test_metric/%d", metric))
		if response == "" {
			return "success"
		}
		return response
	}, cfg.DefaultTimeoutDuration(), 5*time.Second).Should(ContainSubstring("success"))
}

func StartApp(appName string, timeout time.Duration) bool {
	return Expect(cf.Cf("start", appName).Wait(timeout)).To(Exit(0))
}

func CreateTestApp(cfg *config.Config, appType string, initialInstanceCount int) string {
	appName := generator.PrefixedRandomName("autoscaler", appType)
	countStr := strconv.Itoa(initialInstanceCount)
	createApp := cf.Cf("push", appName, "--no-start", "--no-route", "-i", countStr, "-b", cfg.NodejsBuildpackName, "-m", "128M", "-p", config.NODE_APP).Wait(cfg.CfPushTimeoutDuration())
	Expect(createApp).To(Exit(0), "failed creating app")

	mapRouteToApp := cf.Cf("map-route", appName, cfg.AppsDomain, "--hostname", appName).Wait(cfg.DefaultTimeoutDuration())
	Expect(mapRouteToApp).To(Exit(0), "failed to map route to app")
	return appName
}

func DeleteTestApp(appName string, timeout time.Duration) {
	Expect(cf.Cf("delete", appName, "-f", "-r").Wait(timeout)).To(Exit(0))
}

func CurlAppInstance(cfg *config.Config, appName string, appInstance int, url string) string {
	appGuid := GetAppGuid(cfg, appName)
	return cfh.CurlAppWithTimeout(cfg, appName, url, 10*time.Second, "-H", fmt.Sprintf(`X-Cf-App-Instance: %s:%d`, appGuid, appInstance))
}

func AppSetCpuUsage(cfg *config.Config, appName string, percent int, minutes int) {
	Expect(cfh.CurlAppWithTimeout(cfg, appName, fmt.Sprintf("/cpu/%d/%d", percent, minutes), 10*time.Second)).Should(ContainSubstring(`set app cpu utilization`))
}

func AppEndCpuTest(cfg *config.Config, appName string, instance int) {
	Expect(CurlAppInstance(cfg, appName, instance, "/cpu/close")).Should(ContainSubstring(`close cpu test`))
}

func WaitForAppReady(cfg *config.Config, appName string) {
	Eventually(func() string {
		health := &struct {
			Status         string `json:"status"`
			CpuTestRunning bool   `json:"cpuTestRunning"`
		}{}
		err := json.Unmarshal([]byte(cfh.CurlApp(cfg, appName, "/health")), health)
		if err != nil {
			return err.Error()
		}
		return health.Status
	}, cfg.DefaultTimeoutDuration(), 1*time.Second).Should(ContainSubstring("OK"))
}
