package app

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
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
	PolicyPath = "/v1/apps/{appId}/policy"
)

var (
	cfg      *config.Config
	setup    *workflowhelpers.ReproducibleTestSuiteSetup
	interval int
	client   *http.Client
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)

	cfg = config.LoadConfig(t)
	componentName := "Application Scale Suite"
	rs := []Reporter{}

	if cfg.GetArtifactsDirectory() != "" {
		helpers.EnableCFTrace(cfg, componentName)
		rs = append(rs, helpers.NewJUnitReporter(cfg, componentName))
	}

	RunSpecsWithDefaultAndCustomReporters(t, componentName, rs)
}

var _ = BeforeSuite(func() {

	setup = workflowhelpers.NewTestSuiteSetup(cfg)
	setup.Setup()

	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		EnableServiceAccess(cfg, setup.GetOrganizationName())
	})

	serviceExists := cf.Cf("marketplace", "-s", cfg.ServiceName).Wait(cfg.DefaultTimeoutDuration())
	Expect(serviceExists).To(Exit(0), fmt.Sprintf("Service offering, %s, does not exist", cfg.ServiceName))
	interval = cfg.AggregateInterval

	client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
			DisableCompression:  true,
			DisableKeepAlives:   true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.SkipSSLValidation,
			},
		},
		Timeout: 30 * time.Second,
	}

})

var _ = AfterSuite(func() {
	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		DisableServiceAccess(cfg, setup.GetOrganizationName())
	})
	setup.Teardown()
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

func DoAPIRequest(req *http.Request) (*http.Response, error) {
	resp, err := client.Do(req)
	return resp, err
}
