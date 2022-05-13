package healthendpoint_test

import (
	"net/http"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/healthendpoint"
	"code.cloudfoundry.org/lager"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/steinfletcher/apitest"
)

var _ = FDescribe("Health Readiness", func() {

	var (
		t            GinkgoTInterface
		healthServer *mux.Router
		logger       lager.Logger
	)
	const username = "test-user-name"
	const password = "test-user-password"

	BeforeEach(func() {
		t = GinkgoT()
		var err error

		healthServer, err = healthendpoint.HealthBasicAuthRouter(logger, prometheus.NewRegistry(), username, password, "", "")
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("Health endpoint is called without basic auth", func() {
		It("should have json response", func() {
			apitest.New().
				Handler(healthServer).
				Get("/health/readiness").
				Expect(t).
				Status(http.StatusOK).
				Header("Content-Type", "application/json").
				Body(`{ 
	"status" : "OK",
	"checks" : []
}`).
				End()
		})
	})

	Context("Prometheus Health endpoint is called", func() {
		It("should require basic auth", func() {
			apitest.New().
				Handler(healthServer).
				Get("/health").
				Expect(t).
				Status(http.StatusUnauthorized).
				End()
		})
	})

	Context("Health endpoint default response", func() {
		It("should require basic auth", func() {
			apitest.New().
				Handler(healthServer).
				Get("/any").
				Expect(t).
				Status(http.StatusUnauthorized).
				End()
		})
	})

})
