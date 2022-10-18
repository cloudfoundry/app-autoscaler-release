package app_test

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/KevinJCross/cf-test-helpers/v2/helpers"

	"acceptance/config"
	. "acceptance/helpers"

	"github.com/KevinJCross/cf-test-helpers/v2/cf"
	"github.com/KevinJCross/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var (
	cfg      *config.Config
	setup    *workflowhelpers.ReproducibleTestSuiteSetup
	interval int
	client   *http.Client

	instanceName         string
	initialInstanceCount int

	appName string
)

const componentName = "Application Scale Suite"

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, componentName)
}

var _ = BeforeSuite(func() {
	cfg = config.LoadConfig()

	if cfg.GetArtifactsDirectory() != "" {
		helpers.EnableCFTrace(cfg, componentName)
	}

	setup = workflowhelpers.NewTestSuiteSetup(cfg)

	//Cleanup(cfg, setup)

	setup.Setup()

	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		if cfg.IsServiceOfferingEnabled() && cfg.ShouldEnableServiceAccess() {
			EnableServiceAccess(cfg, setup.GetOrganizationName())
		}
	})

	if cfg.IsServiceOfferingEnabled() {
		CheckServiceExists(cfg, setup.TestSpace.SpaceName(), cfg.ServiceName)
	}

	interval = cfg.AggregateInterval

	client = GetHTTPClient(cfg)

})

var _ = AfterSuite(func() {
	if os.Getenv("SKIP_TEARDOWN") == "true" {
		fmt.Println("Skipping Teardown...")
	} else {
		DebugInfo(appName)
		workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
			if cfg.IsServiceOfferingEnabled() && cfg.ShouldEnableServiceAccess() {
				DisableServiceAccess(cfg, setup.GetOrganizationName())
			}
		})
		setup.Teardown()
	}
})

func DebugInfo(anApp string) {
	if os.Getenv("DEBUG") == "true" && cfg.ASApiEndpoint != "" {
		if os.Getenv("CF_PLUGIN_HOME") == "" {
			_ = os.Setenv("CF_PLUGIN_HOME", os.Getenv("HOME"))
		}
		workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
			var commands []*Session
			commands = append(commands, command("cf", "app", anApp))
			commands = append(commands, command("cf", "autoscaling-api", cfg.ASApiEndpoint))
			commands = append(commands, command("cf", "autoscaling-policy", anApp))
			commands = append(commands, command("cf", "autoscaling-history", anApp))
			commands = append(commands, command("cf", "autoscaling-metrics", anApp, "memoryused"))
			commands = append(commands, command("cf", "autoscaling-metrics", anApp, "memoryutil"))
			commands = append(commands, command("cf", "autoscaling-metrics", anApp, "responsetime"))
			commands = append(commands, command("cf", "autoscaling-metrics", anApp, "throughput"))
			commands = append(commands, command("cf", "autoscaling-metrics", anApp, "cpu"))
			commands = append(commands, command("cf", "autoscaling-metrics", anApp, "test_metric"))
			output := new(strings.Builder)
			_, _ = fmt.Fprintf(output, "\n=============== DEBUG ===============\n")
			for _, command := range commands {
				command.Wait(30 * time.Second)
				_, _ = fmt.Fprintf(output, strings.Join(command.Command.Args, " ")+": \n")
				_, _ = fmt.Fprintf(output, string(command.Out.Contents())+"\n")
				_, _ = fmt.Fprintf(output, string(command.Err.Contents())+"\n")
			}
			_, _ = fmt.Fprintf(output, "\n=====================================\n")
			GinkgoWriter.Print(output.String())
		})
	}
}

func command(name string, args ...string) *Session {
	cmd := exec.Command(name, args...)
	start, err := Start(cmd, nil, nil)
	if err != nil {
		GinkgoWriter.Println(err.Error())
	}
	return start
}

func getStartAndEndTime(location *time.Location, offset, duration time.Duration) (time.Time, time.Time) {
	// Since the validation of time could fail if spread over two days and will result in acceptance test failure
	// Need to fix dates in that case.
	startTime := time.Now().In(location).Add(offset)
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
	defer func() { _ = resp.Body.Close() }()
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
