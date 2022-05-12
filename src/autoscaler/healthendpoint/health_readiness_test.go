package healthendpoint_test

import (
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/healthendpoint"
	"code.cloudfoundry.org/lager"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/steinfletcher/apitest"
	"net/http"
)

var _ = FDescribe("Health Readiness", func() {

	var (
		t            GinkgoTInterface
		healthServer *mux.Router
		logger       lager.Logger
	)

	BeforeEach(func() {
		t = GinkgoT()
		var err error
		healthServer, err = healthendpoint.HealthRouter(logger, 999, prometheus.NewRegistry(),
			"test-user-name", "test-user-password", "test-user-hash", "test-password-hash")
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("Health endpoint is called", func() {
		It("should have json response", func() {
			apitest.New().
				Handler(healthServer).
				Get("/health").
				Expect(t).
				Status(http.StatusOK).
				End()
		})
	})

})
