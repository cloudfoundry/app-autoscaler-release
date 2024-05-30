package server_test

import (
	"bytes"
	"net/http"
	"os"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/fakes"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/metricsforwarder/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"

	"testing"
)

func TestServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Server Suite")
}

var (
	rateLimiter *fakes.FakeLimiter

	httpStatusCollector *fakes.FakeHTTPStatusCollector
	httpServer          ifrit.Runner
	serverProcess       ifrit.Process

	conf *config.Config
)

var _ = BeforeSuite(func() {
	_, err := os.ReadFile("../../../../test-certs/metron.key")
	Expect(err).NotTo(HaveOccurred())

	_, err = os.ReadFile("../../../../test-certs/metron.crt")
	Expect(err).NotTo(HaveOccurred())

	_, err = os.ReadFile("../../../../test-certs/loggregator-ca.crt")
	Expect(err).NotTo(HaveOccurred())
})

func CreateRequest(body []byte, path string) *http.Request {
	req, err := http.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	Expect(err).ToNot(HaveOccurred())
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic dXNlcm5hbWU6cGFzc3dvcmQ=")
	return req
}
