package app_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudfoundry-community/go-cfenv"
	"net"
	"net/http"
	"time"

	"code.cloudfoundry.org/app-autoscaler-release/src/acceptance/assets/app/go_app/internal/app"
	"github.com/fgrosse/zaptest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/steinfletcher/apitest"
)

var _ = Describe("Ginkgo/Server", func() {

	var (
		t GinkgoTInterface
	)

	BeforeEach(func() {
		t = GinkgoT()
	})

	Context("basic endpoint tests", func() {
		It("Root should respond correctly", func() {
			apiTest(NoOpSleep, NoOpUseMem, NoOpUseCPU, NoOpPostCustomMetrics).
				Get("/").
				Expect(t).
				Status(http.StatusOK).
				Body(`{"name":"test-app"}`).
				End()
		})
		It("health", func() {
			apiTest(NoOpSleep, NoOpUseMem, NoOpUseCPU, NoOpPostCustomMetrics).
				Get("/health").
				Expect(t).
				Status(http.StatusOK).
				Body(`{"status":"ok"}`).
				End()
		})
	})

	Context("Basic startup", func() {
		var testApp *http.Server
		var client *http.Client
		var port int
		BeforeEach(func() {
			logger := zaptest.LoggerWriter(GinkgoWriter)
			l, err := net.Listen("tcp", ":0")
			Expect(err).ToNot(HaveOccurred())
			port = l.Addr().(*net.TCPAddr).Port
			testApp = app.New(logger, "")
			DeferCleanup(testApp.Close)
			go func() {
				defer GinkgoRecover()
				if err := testApp.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) {
					panic(err)
				}
			}()
			client = &http.Client{Timeout: time.Second * 1}
		})

		It("should start up", func() {
			apitest.New().EnableNetworking(client).Get(fmt.Sprintf("http://localhost:%d/", port)).
				Expect(t).
				Status(http.StatusOK).
				Body(`{"name":"test-app"}`).
				End()
		})

	})
})

func NoOpSleep(_ time.Duration)            {}
func NoOpUseMem(_ uint64)                  {}
func NoOpUseCPU(_ uint64, _ time.Duration) {}
func NoOpPostCustomMetrics(_ context.Context, _ *cfenv.App, _ float64, _ string, _ bool) error {
	return nil
}

func apiTest(sleep func(duration time.Duration), useMem func(useMb uint64), useCPU func(utilization uint64, duration time.Duration), postCustomMetrics func(ctx context.Context, appConfig *cfenv.App, metricsValue float64, metricName string, useMTLS bool) error) *apitest.APITest {
	GinkgoHelper()
	logger := zaptest.LoggerWriter(GinkgoWriter)
	return apitest.New().
		Handler(app.Router(logger, sleep, useMem, useCPU, postCustomMetrics))
}
