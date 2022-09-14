package auth_test

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/app-autoscaler-release/fakes"
	"github.com/cloudfoundry/app-autoscaler-release/helpers"
	"github.com/cloudfoundry/app-autoscaler-release/metricsforwarder/config"
	. "github.com/cloudfoundry/app-autoscaler-release/metricsforwarder/server"
	"github.com/cloudfoundry/app-autoscaler-release/models"

	"code.cloudfoundry.org/lager"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/patrickmn/go-cache"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon_v2"

	"testing"
)

var (
	serverProcess   ifrit.Process
	serverUrl       string
	policyDB        *fakes.FakePolicyDB
	rateLimiter     *fakes.FakeLimiter
	fakeCredentials *fakes.FakeCredentials

	credentialCache    cache.Cache
	allowedMetricCache cache.Cache
)

func TestAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Auth Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {

	_, err := os.ReadFile("../../../../../test-certs/metron.key")
	Expect(err).NotTo(HaveOccurred())

	_, err = os.ReadFile("../../../../../test-certs/metron.crt")
	Expect(err).NotTo(HaveOccurred())

	_, err = os.ReadFile("../../../../../test-certs/loggregator-ca.crt")
	Expect(err).NotTo(HaveOccurred())

	return nil
}, func(_ []byte) {

	testCertDir := "../../../../../test-certs"
	loggregatorConfig := config.LoggregatorConfig{
		TLS: models.TLSCerts{
			KeyFile:    filepath.Join(testCertDir, "metron.key"),
			CertFile:   filepath.Join(testCertDir, "metron.crt"),
			CACertFile: filepath.Join(testCertDir, "loggregator-ca.crt"),
		},
	}
	serverConfig := config.ServerConfig{
		Port: 2222 + GinkgoParallelProcess(),
	}

	loggerConfig := helpers.LoggingConfig{
		Level: "debug",
	}

	conf := &config.Config{
		Server:            serverConfig,
		Logging:           loggerConfig,
		LoggregatorConfig: loggregatorConfig,
	}
	policyDB = &fakes.FakePolicyDB{}
	credentialCache = *cache.New(10*time.Minute, -1)
	allowedMetricCache = *cache.New(10*time.Minute, -1)
	httpStatusCollector := &fakes.FakeHTTPStatusCollector{}
	rateLimiter = &fakes.FakeLimiter{}
	fakeCredentials = &fakes.FakeCredentials{}

	httpServer, err := NewServer(lager.NewLogger("test"), conf, policyDB,
		fakeCredentials, allowedMetricCache, httpStatusCollector, rateLimiter)
	Expect(err).NotTo(HaveOccurred())
	serverUrl = fmt.Sprintf("http://127.0.0.1:%d", conf.Server.Port)
	serverProcess = ginkgomon_v2.Invoke(httpServer)
})

var _ = SynchronizedAfterSuite(func() {
	ginkgomon_v2.Interrupt(serverProcess)
}, func() {
})
