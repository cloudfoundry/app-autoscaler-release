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

	Context("Health endpoint is called", func() {
		It("should have json response", func() {
			apitest.New().
				Handler(healthServer).
				Get("/health/readiness").
				BasicAuth(username, password).
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

})
