package app_test

import (
	"acceptance/assets/app/go_app/internal/app"
	"errors"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
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
			apiTest(NoOpSleep, NoOpUseMem).
				Get("/").
				Expect(t).
				Status(http.StatusOK).
				Body(`{"name":"test-app"}`).
				End()
		})
		It("health", func() {
			apiTest(NoOpSleep, NoOpUseMem).
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
		BeforeEach(func() {
			logger := logrus.New()
			testApp = app.New(logger, "localhost:31253")
			DeferCleanup(testApp.Close)
			go func() {
				defer GinkgoRecover()
				if err := testApp.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					panic(err)
				}
			}()
			client = &http.Client{Timeout: time.Second * 1}
		})

		It("should start up", func() {
			apitest.New().EnableNetworking(client).Get("http://localhost:31253/").
				Expect(t).
				Status(http.StatusOK).
				Body(`{"name":"test-app"}`).
				End()
		})

	})
})

func NoOpSleep(_ time.Duration) {}
func NoOpUseMem(_ uint64)       {}

func apiTest(sleep func(duration time.Duration), useMem func(useMb uint64)) *apitest.APITest {
	return apitest.New().
		Handler(app.Router(logrus.New(), sleep, useMem))
}
