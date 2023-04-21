package app_test

import (
	"code.cloudfoundry.org/app-autoscaler-release/src/acceptance/assets/app/go_app/internal/app"
	api "code.cloudfoundry.org/app-autoscaler-release/src/acceptance/assets/app/go_app/internal/custommetrics"
	"context"
	"github.com/cloudfoundry-community/go-cfenv"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"net/http"
	"sync"
)

var _ = Describe("custom metrics tests", func() {

	Context("custom metrics handler", func() {
		var (
			sentValue  float64
			sentMetric string
			mtlsUsed   bool
		)
		customMetricsLock := &sync.Mutex{}
		customMetricsLock.Lock()
		postCustomMetrics := func(ctx context.Context, appConfig *cfenv.App, metricsValue float64, metricName string, useMtls bool) error {
			sentValue = metricsValue
			sentMetric = metricName
			mtlsUsed = useMtls
			customMetricsLock.Unlock()
			return nil
		}

		It("should err if value out of bounds", func() {
			apiTest(NoOpSleep, NoOpUseMem, NoOpUseCPU, NoOpPostCustomMetrics).
				Get("/custom-metrics/test/100001010101010249032897287298719874687936483275648273632429479827398798271").
				Expect(GinkgoT()).
				Status(http.StatusBadRequest).
				Body(`{"error":{"description":"invalid metric value: strconv.ParseUint: parsing \"100001010101010249032897287298719874687936483275648273632429479827398798271\": value out of range"}}`).
				End()
		})
		It("should err if value not a number", func() {
			apiTest(NoOpSleep, NoOpUseMem, NoOpUseCPU, NoOpPostCustomMetrics).
				Get("/custom-metrics/test/invalid").
				Expect(GinkgoT()).
				Status(http.StatusBadRequest).
				Body(`{"error":{"description":"invalid metric value: strconv.ParseUint: parsing \"invalid\": invalid syntax"}}`).
				End()
		})
		It("should post the custom metric", func() {
			apiTest(NoOpSleep, NoOpUseMem, NoOpUseCPU, postCustomMetrics).
				Get("/custom-metrics/test/4").
				Expect(GinkgoT()).
				Status(http.StatusOK).
				Body(`{"mtls":false}`).
				End()
			customMetricsLock.Lock()
			Expect(sentMetric).Should(Equal("test"))
			Expect(sentValue).Should(Equal(4.0))
			Expect(mtlsUsed).Should(Equal(false))
		})
	})
	Context("PostCustomMetrics", func() {
		It("should post a custom metric", func() {

			testAppId := "test-app-id"
			fakeServer := ghttp.NewServer()
			fakeServer.AppendHandlers(
				ghttp.VerifyRequest("POST", "/v1/apps/"+testAppId+"/metrics"),
			)

			customMetricsCredentials := api.CustomMetricsCredentials{
				Username: "user",
				Password: "pass",
				URL:      fakeServer.URL(),
			}
			service := cfenv.Service{
				Name:        "test",
				Tags:        []string{"app-autoscaler"},
				Credentials: map[string]interface{}{"custom_metrics": customMetricsCredentials},
			}

			appEnv := cfenv.App{
				AppID:    testAppId,
				Index:    0,
				Services: map[string][]cfenv.Service{"autoscaler": {service}},
			}

			err := app.PostCustomMetric(context.TODO(), &appEnv, 42, "test", false)
			Expect(err).ToNot(HaveOccurred())

			Expect(len(fakeServer.ReceivedRequests())).To(Equal(1))
			fakeServer.Close()
		})
	})
})
