package helpers

import (
	"acceptance/app"
	"acceptance/config"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

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
	curler := app.NewAppCurler(cfg)
	Eventually(func() string {
		response := curler.Curl(appName, fmt.Sprintf("/custom-metrics/test_metric/%d", metric), 60*time.Second)
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
