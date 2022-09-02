package helpers

import (
	"acceptance/config"
	"encoding/json"
	"fmt"
	"github.com/onsi/ginkgo/v2"
	"strconv"
	"time"

	"github.com/KevinJCross/cf-test-helpers/v2/cf"
	"github.com/KevinJCross/cf-test-helpers/v2/generator"
	cfh "github.com/KevinJCross/cf-test-helpers/v2/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func GetApps(cfg *config.Config, orgGuid, spaceGuid string, prefix string) []string {
	rawApps := getRawApps(spaceGuid , orgGuid , cfg.DefaultTimeoutDuration())
	return filterByPrefix(prefix, getNames(rawApps))
}

func getRawAppsByPage(spaceGuid string, orgGuid string, page int, timeout time.Duration) cfResourceObject {
	var appsResponse cfResourceObject
	rawApps := cf.Cf("curl", "/v3/apps?space_guids="+spaceGuid+"&organization_guids="+orgGuid+"&page="+strconv.Itoa(page)).Wait(timeout)
	Expect(rawApps).To(Exit(0), "unable to get apps")
	err := json.Unmarshal(rawApps.Out.Contents(), &appsResponse)
	Expect(err).ShouldNot(HaveOccurred())
	return appsResponse
}

func getRawApps(spaceGuid string, orgGuid string, timeout time.Duration) []cfResource {
	var rawApps []cfResource
	totalPages := 1

	for page := 1; page <= totalPages; page++ {
		var appsResponse = getRawAppsByPage(spaceGuid, orgGuid, page, timeout)
		GinkgoWriter.Println(appsResponse.Pagination.TotalPages)
		totalPages = appsResponse.Pagination.TotalPages
		rawApps = append(rawApps, appsResponse.Resources... )
	}

	return rawApps
}

func GetRunningApps(cfg *config.Config, orgGuid, spaceGuid string) []string {
	rawApps := getRawApps(spaceGuid , orgGuid , cfg.DefaultTimeoutDuration())
	return filterByState(rawApps, "RUNNING")
}


func DeleteApps(cfg *config.Config, apps []string, threshold int) {
	for _, app := range apps {
		deleteApp := cf.Cf("delete", app, "-f").Wait(cfg.DefaultTimeoutDuration())
		Expect(deleteApp).To(Exit(0), fmt.Sprintf("unable to delete app %s", app))
	}
}

func SendMetric(cfg *config.Config, appName string, metric int) {
	cfh.CurlApp(cfg, appName, fmt.Sprintf("/custom-metrics/test_metric/%d", metric), "-f")
}

func StartApp(appName string, timeout time.Duration) bool {
	startApp := cf.Cf("start", appName).Wait(timeout)
	if startApp.ExitCode() != 0 {
		cf.Cf("logs", appName, "--recent").Wait(2 * time.Minute)
	}
	return Expect(startApp).To(Exit(0))
}

func CreateTestApp(cfg *config.Config, appType string, initialInstanceCount int) string {
	appName := generator.PrefixedRandomName(cfg.Prefix, appType)
	By("Creating test app")
	CreateTestAppByName(*cfg, appName, initialInstanceCount)
	return appName
}

func CreateTestAppByName(cfg config.Config, appName string, initialInstanceCount int) {

	setNodeTLSRejectUnauthorizedEnvironmentVariable := "1"
	if cfg.GetSkipSSLValidation() {
		setNodeTLSRejectUnauthorizedEnvironmentVariable = "0"
	}

	countStr := strconv.Itoa(initialInstanceCount)
	createApp := cf.Cf("push",
		"--var", "app_name="+appName,
		"--var", "app_domain="+cfg.AppsDomain,
		"--var", "service_name="+cfg.ServiceName,
		"--var", "instances="+countStr,
		"--var", "buildpack="+cfg.NodejsBuildpackName,
		"--var", "node_tls_reject_unauthorized="+setNodeTLSRejectUnauthorizedEnvironmentVariable,
		"-p", config.NODE_APP,
		"-f", config.NODE_APP+"/app_manifest.yml",
		"--no-start",
	).Wait(cfg.CfPushTimeoutDuration())


	if createApp.ExitCode() != 0 {
		cf.Cf("logs", appName, "--recent").Wait(2 * time.Minute)
	}
	Expect(createApp).To(Exit(0), "failed creating app")

	ginkgo.GinkgoWriter.Printf("\nfinish creating test app: %s\n", appName)
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
