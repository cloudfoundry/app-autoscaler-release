package server_test

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/fakes"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/metricsforwarder/config"
	. "code.cloudfoundry.org/app-autoscaler/src/autoscaler/metricsforwarder/server"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"

	"code.cloudfoundry.org/lager/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit/ginkgomon_v2"

	"net/http"
	"net/http/httptest"

	"github.com/patrickmn/go-cache"
)

var _ = Describe("MetricHandler", func() {
	var (
		fakeCredentials *fakes.FakeCredentials

		policyDB *fakes.FakePolicyDB
		handler  *CustomMetricsHandler

		allowedMetricCache cache.Cache

		allowedMetricTypeSet map[string]struct{}

		metricsforwarder *fakes.FakeMetricForwarder

		resp *httptest.ResponseRecorder
		req  *http.Request
		err  error
		body []byte

		vars map[string]string

		found bool

		scalingPolicy *models.ScalingPolicy
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
		serverConfig := helpers.ServerConfig{
			Port: 2223 + GinkgoParallelProcess(),
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
		allowedMetricCache = *cache.New(10*time.Minute, -1)
		httpStatusCollector := &fakes.FakeHTTPStatusCollector{}
		rateLimiter = &fakes.FakeLimiter{}
		fakeCredentials = &fakes.FakeCredentials{}

		logger := lager.NewLogger("metrichandler-test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))

		httpServer, err = NewServer(logger, conf, policyDB,
			fakeCredentials, allowedMetricCache, httpStatusCollector, rateLimiter)
		Expect(err).NotTo(HaveOccurred())
		serverUrl = fmt.Sprintf("http://127.0.0.1:%d", conf.Server.Port)
		serverProcess = ginkgomon_v2.Invoke(httpServer)

		policyDB = &fakes.FakePolicyDB{}
		metricsforwarder = &fakes.FakeMetricForwarder{}
		allowedMetricCache = *cache.New(10*time.Minute, -1)
		allowedMetricTypeSet = make(map[string]struct{})
		vars = make(map[string]string)
		resp = httptest.NewRecorder()
		handler = NewCustomMetricsHandler(logger, metricsforwarder, policyDB, allowedMetricCache)
		allowedMetricCache.Flush()
	})

	JustBeforeEach(func() {

	})
	AfterEach(func() {
		ginkgomon_v2.Interrupt(serverProcess)
		// wait for the server to shutdown
		// this is necessary because the server is running in a separate goroutine
		// and the test can exit before the server has a chance to shutdown

		Eventually(serverProcess.Wait(), 5).Should(Receive())

	})

	JustBeforeEach(func() {
		req = CreateRequest(body, serverUrl+"/v1/apps/an-app-id/metrics")
		req.SetBasicAuth("username", "password")
		vars["appid"] = "an-app-id"
		handler.VerifyCredentialsAndPublishMetrics(resp, req, vars)
	})

	Describe("PublishMetrics", func() {

		Context("when a request to publish custom metrics comes with malformed request body", func() {
			BeforeEach(func() {
				policyDB.GetCredentialReturns(&models.Credential{
					Username: "$2a$10$YnQNQYcvl/Q2BKtThOKFZ.KB0nTIZwhKr5q1pWTTwC/PUAHsbcpFu",
					Password: "$2a$10$6nZ73cm7IV26wxRnmm5E1.nbk9G.0a4MrbzBFPChkm5fPftsUwj9G",
				}, nil)
				body = []byte(`{
					   "instance_index":0,
					   "test" :
					   "metrics":[
					      {
					         "name":"custom_metric1",
					         "type":"gauge",
					         "value":200,
					         "unit":"unit"
					      }
					   ]
				}`)
			})

			It("returns status code 400", func() {
				Expect(resp.Code).To(Equal(http.StatusBadRequest))
				errJson := &models.ErrorResponse{}
				err = json.Unmarshal(resp.Body.Bytes(), errJson)
				Expect(errJson).To(Equal(&models.ErrorResponse{
					Code:    "Bad-Request",
					Message: "Error unmarshaling custom metrics request body",
				}))
			})

		})

		Context("when a valid request to publish custom metrics comes", func() {
			Context("when allowedMetrics exists in the cache", func() {
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
					allowedMetricTypeSet["queuelength"] = struct{}{}
					allowedMetricCache.Set("an-app-id", allowedMetricTypeSet, 10*time.Minute)
					customMetrics := []*models.CustomMetric{
						{
							Name: "queuelength", Value: 12, Unit: "unit", InstanceIndex: 1, AppGUID: "an-app-id",
						},
					}
					body, err = json.Marshal(models.MetricsConsumer{InstanceIndex: 0, CustomMetrics: customMetrics})
					Expect(err).NotTo(HaveOccurred())
				})

				It("should get the allowedMetrics from cache without searching from database and returns status code 200", func() {
					Expect(policyDB.GetAppPolicyCallCount()).To(Equal(0))
					Expect(resp.Code).To(Equal(http.StatusOK))
				})

			})

			Context("when allowedMetrics does not exists in the cache but exist in the database", func() {
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
					customMetrics := []*models.CustomMetric{
						{
							Name: "queuelength", Value: 12, Unit: "unit", InstanceIndex: 1, AppGUID: "an-app-id",
						},
					}
					body, err = json.Marshal(models.MetricsConsumer{InstanceIndex: 0, CustomMetrics: customMetrics})
					Expect(err).NotTo(HaveOccurred())
				})

				It("should get the allowedMetrics from database and add it to the cache and returns status code 200", func() {
					Expect(policyDB.GetAppPolicyCallCount()).To(Equal(1))
					Expect(resp.Code).To(Equal(http.StatusOK))
					_, found = allowedMetricCache.Get("an-app-id")
					Expect(found).To(Equal(true))
				})

			})

			Context("when allowedMetrics neither exists in the cache nor exist in the database", func() {
				BeforeEach(func() {
					customMetrics := []*models.CustomMetric{
						{
							Name: "queuelength", Value: 12, Unit: "unit", InstanceIndex: 1, AppGUID: "an-app-id",
						},
					}
					body, err = json.Marshal(models.MetricsConsumer{InstanceIndex: 0, CustomMetrics: customMetrics})
					Expect(err).NotTo(HaveOccurred())
				})

				It("should search in both cache & database and returns status code 400", func() {
					Expect(policyDB.GetAppPolicyCallCount()).To(Equal(1))
					Expect(resp.Code).To(Equal(http.StatusBadRequest))
					errJson := &models.ErrorResponse{}
					err = json.Unmarshal(resp.Body.Bytes(), errJson)
					Expect(errJson).To(Equal(&models.ErrorResponse{
						Code:    "Bad-Request",
						Message: "no policy defined",
					}))
				})
			})
		})

		Context("when a request to publish custom metrics comes with standard metric type", func() {
			BeforeEach(func() {
				policyDB.GetCredentialReturns(&models.Credential{
					Username: "$2a$10$YnQNQYcvl/Q2BKtThOKFZ.KB0nTIZwhKr5q1pWTTwC/PUAHsbcpFu",
					Password: "$2a$10$6nZ73cm7IV26wxRnmm5E1.nbk9G.0a4MrbzBFPChkm5fPftsUwj9G",
				}, nil)
				body = []byte(`{
					   "instance_index":0,
					   "metrics":[
					      {
					         "name":"memoryused",
					         "type":"gauge",
					         "value":200,
					         "unit":"unit"
					      }
					   ]
				}`)
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
			})

			It("returns status code 400", func() {
				Expect(resp.Code).To(Equal(http.StatusBadRequest))
				errJson := &models.ErrorResponse{}
				err = json.Unmarshal(resp.Body.Bytes(), errJson)
				Expect(errJson).To(Equal(&models.ErrorResponse{
					Code:    "Bad-Request",
					Message: "Custom Metric: memoryused matches with standard metrics name",
				}))
			})

		})

		Context("when a request to publish custom metrics comes with non allowed metric types", func() {
			BeforeEach(func() {
				policyDB.GetCredentialReturns(&models.Credential{
					Username: "$2a$10$YnQNQYcvl/Q2BKtThOKFZ.KB0nTIZwhKr5q1pWTTwC/PUAHsbcpFu",
					Password: "$2a$10$6nZ73cm7IV26wxRnmm5E1.nbk9G.0a4MrbzBFPChkm5fPftsUwj9G",
				}, nil)
				body = []byte(`{
					   "instance_index":0,
					   "metrics":[
					      {
					         "name":"wrong_metric_type",
					         "type":"gauge",
					         "value":200,
					         "unit":"unit"
					      }
					   ]
				}`)
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
			})

			It("returns status code 400", func() {
				Expect(resp.Code).To(Equal(http.StatusBadRequest))
				errJson := &models.ErrorResponse{}
				err = json.Unmarshal(resp.Body.Bytes(), errJson)
				Expect(errJson).To(Equal(&models.ErrorResponse{
					Code:    "Bad-Request",
					Message: "Custom Metric: wrong_metric_type does not match with metrics defined in policy",
				}))
			})

		})
	})

})
