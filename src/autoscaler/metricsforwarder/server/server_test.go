package server_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/fakes"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/metricsforwarder/config"
	. "code.cloudfoundry.org/app-autoscaler/src/autoscaler/metricsforwarder/server"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"
	"code.cloudfoundry.org/lager/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/patrickmn/go-cache"
	"github.com/tedsuo/ifrit/ginkgomon_v2"
)

var _ = Describe("CustomMetrics Server", func() {

	var (
		policyDB        *fakes.FakePolicyDB
		fakeCredentials *fakes.FakeCredentials

		resp          *http.Response
		req           *http.Request
		body          []byte
		err           error
		scalingPolicy *models.ScalingPolicy

		allowedMetricCache cache.Cache

		customMetrics []*models.CustomMetric
		serverUrl     string
	)

	BeforeEach(func() {
		testCertDir := "../../../../test-certs"

		loggregatorConfig := config.LoggregatorConfig{
			TLS: models.TLSCerts{
				KeyFile:    filepath.Join(testCertDir, "metron.key"),
				CertFile:   filepath.Join(testCertDir, "metron.crt"),
				CACertFile: filepath.Join(testCertDir, "loggregator-ca.crt"),
			},
			MetronAddress: "invalid-host-name-blah:12345",
		}

		//random number from 1 to 1000
		//to avoid port conflict in parallel test

		rand.Seed(time.Now().UnixNano())
		randomNumber := rand.Intn(1000) + 1

		serverConfig := helpers.ServerConfig{
			Port: 2222 + randomNumber,
		}

		fmt.Printf("serverConfig.Port: %d \n", serverConfig.Port)

		loggerConfig := helpers.LoggingConfig{
			Level: "debug",
		}

		conf = &config.Config{
			Server:            serverConfig,
			Logging:           loggerConfig,
			LoggregatorConfig: loggregatorConfig,
		}
		policyDB = &fakes.FakePolicyDB{}
		fakeCredentials = &fakes.FakeCredentials{}

		allowedMetricCache = *cache.New(10*time.Minute, -1)
		httpStatusCollector = &fakes.FakeHTTPStatusCollector{}
		rateLimiter = &fakes.FakeLimiter{}

		customMetrics = []*models.CustomMetric{
			{
				Name: "queuelength", Value: 12, Unit: "unit", InstanceIndex: 1, AppGUID: "an-app-id",
			},
		}
	})

	JustBeforeEach(func() {
		logger := lager.NewLogger("server_suite_test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		fmt.Printf("credentials 2: %p \n", &fakeCredentials)
		httpServer, err = NewServer(logger, conf, policyDB,
			fakeCredentials, allowedMetricCache, httpStatusCollector, rateLimiter)
		Expect(err).NotTo(HaveOccurred())
		serverUrl = fmt.Sprintf("http://127.0.0.1:%d", conf.Server.Port)
		serverProcess = ginkgomon_v2.Invoke(httpServer)
	})

	AfterEach(func() {
		ginkgomon_v2.Interrupt(serverProcess)
		fakeCredentials = nil
		serverUrl = ""
	})

	Context("when a request to forward custom metrics comes", func() {
		BeforeEach(func() {
			scalingPolicy = &models.ScalingPolicy{
				InstanceMin: 1,
				InstanceMax: 6,
				ScalingRules: []*models.ScalingRule{{
					MetricType:            "queuelength",
					BreachDurationSeconds: 60,
					Threshold:             10,
					Operator:              ">",
					CoolDownSeconds:       60,
					Adjustment:            "+1"}}}
			policyDB.GetAppPolicyReturns(scalingPolicy, nil)
			fakeCredentials.ValidateReturns(true, nil)

		})

		It("returns status code 200", func() {
			Expect(fakeCredentials.ValidateCallCount()).To(Equal(0))
			body, err = json.Marshal(models.MetricsConsumer{InstanceIndex: 0, CustomMetrics: customMetrics})
			Expect(err).NotTo(HaveOccurred())

			client := &http.Client{}

			req = CreateRequest(body, serverUrl+"/v1/apps/san-app-id/metrics")
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			resp.Body.Close()
		})
	})

	Context("when a request to forward custom metrics comes without Authorization header", func() {
		BeforeEach(func() {
			fakeCredentials.ValidateReturns(false, nil)
		})

		It("returns status code 401", func() {
			Expect(fakeCredentials.ValidateCallCount()).To(Equal(0))
			body, err = json.Marshal(models.MetricsConsumer{InstanceIndex: 0, CustomMetrics: customMetrics})
			Expect(err).NotTo(HaveOccurred())
			client := &http.Client{}
			req = CreateRequest(body, serverUrl+"/v1/apps/san-app-id/metrics")
			resp, err = client.Do(req)
			Expect(fakeCredentials.ValidateCallCount()).To(Equal(1))
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			resp.Body.Close()
		})
	})

	Context("when a request to forward custom metrics comes without 'Basic'", func() {
		BeforeEach(func() {
			fakeCredentials.ValidateReturns(true, nil)
		})

		It("returns status code 401", func() {
			Expect(fakeCredentials.ValidateCallCount()).To(Equal(0))
			body, err = json.Marshal(models.CustomMetric{Name: "queuelength", Value: 12, Unit: "unit", InstanceIndex: 123, AppGUID: "an-app-id"})
			Expect(err).NotTo(HaveOccurred())
			req = CreateRequest(body, serverUrl+"/v1/apps/san-app-id/metrics")
			client := &http.Client{}
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCredentials.ValidateCallCount()).To(Equal(1))
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			resp.Body.Close()
		})
	})

	Context("when a request to forward custom metrics comes with wrong user credentials", func() {
		BeforeEach(func() {
			fakeCredentials.ValidateReturns(false, errors.New("wrong credentials"))
		})

		It("returns status code 401", func() {
			Expect(fakeCredentials.ValidateCallCount()).To(Equal(0))
			body, err = json.Marshal(models.CustomMetric{Name: "queuelength", Value: 12, Unit: "unit", InstanceIndex: 123, AppGUID: "an-app-id"})
			Expect(err).NotTo(HaveOccurred())

			client := &http.Client{}
			req = CreateRequest(body, serverUrl+"/v1/apps/an-app-id/metrics")
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCredentials.ValidateCallCount()).To(Equal(1))
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			resp.Body.Close()
		})
	})

	Context("when a request to forward custom metrics comes with unmatched metric types", func() {
		BeforeEach(func() {
			fakeCredentials.ValidateReturns(true, nil)
			fmt.Printf("credentials 1: %p \n", &fakeCredentials)

		})

		It("returns status code 400", func() {
			fmt.Printf("credentials 3: %p \n", &fakeCredentials)
			Expect(fakeCredentials.ValidateCallCount()).To(Equal(0))
			body, err = json.Marshal(models.CustomMetric{Name: "queuelength", Value: 12, Unit: "unit", InstanceIndex: 123, AppGUID: "an-app-id"})
			Expect(err).NotTo(HaveOccurred())
			client := &http.Client{}
			req = CreateRequest(body, serverUrl+"/v1/apps/an-app-id/metrics")
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCredentials.ValidateCallCount()).To(Equal(1))
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			resp.Body.Close()
		})
	})

	Context("when multiple requests to forward custom metrics comes beyond ratelimit", func() {
		BeforeEach(func() {
			rateLimiter.ExceedsLimitReturns(true)
			scalingPolicy = &models.ScalingPolicy{
				InstanceMin: 1,
				InstanceMax: 6,
				ScalingRules: []*models.ScalingRule{{
					MetricType:            "queuelength",
					BreachDurationSeconds: 60,
					Threshold:             10,
					Operator:              ">",
					CoolDownSeconds:       60,
					Adjustment:            "+1"}}}

			policyDB.GetAppPolicyReturns(scalingPolicy, nil)

			fakeCredentials.ValidateReturns(true, nil)
		})

		It("returns status code 429", func() {
			Expect(fakeCredentials.ValidateCallCount()).To(Equal(0))
			body, err = json.Marshal(models.MetricsConsumer{InstanceIndex: 0, CustomMetrics: customMetrics})
			Expect(err).NotTo(HaveOccurred())

			client := &http.Client{}
			req = CreateRequest(body, serverUrl+"/v1/apps/an-app-id/metrics")
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusTooManyRequests))
			resp.Body.Close()
		})
	})
})
