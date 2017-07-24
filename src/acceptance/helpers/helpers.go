package helpers

import (
	"acceptance/config"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func Curl(cfg *config.Config, args ...string) (int, []byte, error) {
	curlCmd := helpers.Curl(cfg, append([]string{"--output", "/dev/stderr", "--write-out", "%{http_code}"}, args...)...).Wait(cfg.DefaultTimeoutDuration())
	if curlCmd.ExitCode() != 0 {
		return 0, curlCmd.Err.Contents(), fmt.Errorf("curl failed: exit code %d", curlCmd.ExitCode())
	}
	statusCode, err := strconv.Atoi(string(curlCmd.Out.Contents()))
	if err != nil {
		return 0, curlCmd.Err.Contents(), err
	}
	return statusCode, curlCmd.Err.Contents(), nil
}

func OauthToken(cfg *config.Config) string {
	cmd := cf.Cf("oauth-token")
	Expect(cmd.Wait(cfg.DefaultTimeoutDuration())).To(Exit(0))
	return strings.TrimSpace(string(cmd.Out.Contents()))
}

func EnableServiceAccess(cfg *config.Config, orgName string) {
	enableServiceAccess := cf.Cf("enable-service-access", cfg.ServiceName, "-o", orgName).Wait(cfg.DefaultTimeoutDuration())
	Expect(enableServiceAccess).To(Exit(0), fmt.Sprintf("Failed to enable service %s for org %s", cfg.ServiceName, orgName))
}

func DisableServiceAccess(cfg *config.Config, orgName string) {
	enableServiceAccess := cf.Cf("disable-service-access", cfg.ServiceName, "-o", orgName).Wait(cfg.DefaultTimeoutDuration())
	Expect(enableServiceAccess).To(Exit(0), fmt.Sprintf("Failed to disable service %s for org %s", cfg.ServiceName, orgName))
}
