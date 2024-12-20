package integration_test

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/testhelpers"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Integration_GolangApi_EventGenerator", func() {
	var (
		appId             string
		pathVariables     []string
		parameters        map[string]string
		metric            *models.AppMetric
		metricType        = "memoryused"
		initInstanceCount = 2
		serviceInstanceId string
		bindingId         string
		orgId             string
		spaceId           string
	)

	BeforeEach(func() {
		startFakeCCNOAAUAA(initInstanceCount)
		httpClient = testhelpers.NewApiClient()
		httpClientForPublicApi = testhelpers.NewPublicApiClient()

		eventGeneratorConfPath = components.PrepareEventGeneratorConfig(dbUrl, components.Ports[EventGenerator], fmt.Sprintf("https://127.0.0.1:%d", components.Ports[MetricsCollector]), fmt.Sprintf("https://127.0.0.1:%d", components.Ports[ScalingEngine]), aggregatorExecuteInterval, policyPollerInterval, saveInterval, evaluationManagerInterval, defaultHttpClientTimeout, tmpDir)
		startEventGenerator()
		golangApiServerConfPath = components.PrepareGolangApiServerConfig(
			dbUrl,
			components.Ports[GolangAPIServer],
			components.Ports[GolangServiceBroker],
			fakeCCNOAAUAA.URL(),
			fmt.Sprintf("https://127.0.0.1:%d", components.Ports[Scheduler]),
			fmt.Sprintf("https://127.0.0.1:%d", components.Ports[ScalingEngine]),
			fmt.Sprintf("https://127.0.0.1:%d", components.Ports[EventGenerator]),
			"https://127.0.0.1:8888",
			tmpDir)
		brokerAuth = base64.StdEncoding.EncodeToString([]byte("broker_username:broker_password"))
		startGolangApiServer()
		serviceInstanceId = getRandomIdRef("serviceInstId")
		orgId = getRandomIdRef("orgId")
		spaceId = getRandomIdRef("spaceId")
		bindingId = getRandomIdRef("bindingId")
		appId = getRandomIdRef("appId")
		pathVariables = []string{appId, metricType}

	})

	AfterEach(func() {
		stopGolangApiServer()
		stopEventGenerator()
	})
	Describe("Get App Metrics", func() {

		Context("Cloud Controller api is not available", func() {
			BeforeEach(func() {
				fakeCCNOAAUAA.Reset()
				fakeCCNOAAUAA.AllowUnhandledRequests = true
			})
			It("should error with status code 500", func() {
				checkPublicAPIResponseContentWithParameters(getAppAggregatedMetrics, components.Ports[GolangAPIServer], pathVariables, parameters, http.StatusInternalServerError, map[string]interface{}{
					"code":    "Internal-Server-Error",
					"message": "Failed to check if user is admin",
				})
			})
		})

		Context("UAA api is not available", func() {
			BeforeEach(func() {
				fakeCCNOAAUAA.Reset()
				fakeCCNOAAUAA.AllowUnhandledRequests = true
				fakeCCNOAAUAA.Add().Info(fakeCCNOAAUAA.URL())
			})
			It("should error with status code 500", func() {
				checkPublicAPIResponseContentWithParameters(getAppAggregatedMetrics, components.Ports[GolangAPIServer], pathVariables, parameters, http.StatusInternalServerError, map[string]interface{}{
					"code":    "Internal-Server-Error",
					"message": "Failed to check if user is admin",
				})
			})
		})
		Context("UAA api returns 401", func() {
			BeforeEach(func() {
				fakeCCNOAAUAA.Reset()
				fakeCCNOAAUAA.AllowUnhandledRequests = true
				fakeCCNOAAUAA.Add().Info(fakeCCNOAAUAA.URL()).Introspect(testUserScope).UserInfo(http.StatusUnauthorized, "ERR")
			})
			It("should error with status code 401", func() {
				checkPublicAPIResponseContentWithParameters(getAppAggregatedMetrics, components.Ports[GolangAPIServer], pathVariables,
					parameters, http.StatusUnauthorized, map[string]interface{}{
						"code":    "Unauthorized",
						"message": "You are not authorized to perform the requested action"})
			})
		})

		Context("Check permission not passed", func() {
			BeforeEach(func() {
				fakeCCNOAAUAA.Add().Roles(http.StatusOK)
			})
			It("should error with status code 401", func() {
				checkPublicAPIResponseContentWithParameters(getAppAggregatedMetrics, components.Ports[GolangAPIServer],
					pathVariables, parameters, http.StatusUnauthorized, map[string]interface{}{
						"code":    "Unauthorized",
						"message": "You are not authorized to perform the requested action",
					})
			})
		})

		When("the app is bound to the service instance", func() {
			BeforeEach(func() {
				provisionAndBind(serviceInstanceId, orgId, spaceId, bindingId, appId, components.Ports[GolangServiceBroker], httpClientForPublicApi)
			})

			Context("EventGenerator is down", func() {
				JustBeforeEach(func() {
					stopEventGenerator()
				})

				It("should error with status code 500", func() {
					checkPublicAPIResponseContentWithParameters(getAppAggregatedMetrics, components.Ports[GolangAPIServer], pathVariables, parameters, http.StatusInternalServerError, map[string]interface{}{
						"code":    "Internal Server Error",
						"message": "Error retrieving metrics history from eventgenerator",
					})
				})
			})

			Context("Get aggregated metrics", func() {
				BeforeEach(func() {
					metric = &models.AppMetric{
						AppId:      appId,
						MetricType: models.MetricNameMemoryUsed,
						Unit:       models.UnitMegaBytes,
						Value:      "123456",
					}

					metric.Timestamp = 666666
					insertAppMetric(metric)

					metric.Timestamp = 555555
					insertAppMetric(metric)

					metric.Timestamp = 555556
					insertAppMetric(metric)

					metric.Timestamp = 333333
					insertAppMetric(metric)

					metric.Timestamp = 444444
					insertAppMetric(metric)

					//add some other metric-type
					metric.MetricType = models.MetricNameThroughput
					metric.Unit = models.UnitNum
					metric.Timestamp = 444444
					insertAppMetric(metric)
					//add some  other appId
					metric.AppId = getRandomIdRef("metric.appId")
					metric.MetricType = models.MetricNameMemoryUsed
					metric.Unit = models.UnitMegaBytes
					metric.Timestamp = 444444
					insertAppMetric(metric)
				})
				It("should get the metrics ", func() {
					By("get the 1st page")
					parameters = map[string]string{"start-time": "111111", "end-time": "999999", "order-direction": "asc", "page": "1", "results-per-page": "2"}
					result := AppAggregatedMetricResult{
						TotalResults: 5,
						TotalPages:   3,
						Page:         1,
						NextUrl:      getAppAggregatedMetricUrl(appId, metricType, parameters, 2),
						Resources: []models.AppMetric{
							{
								AppId:      appId,
								MetricType: models.MetricNameMemoryUsed,
								Unit:       models.UnitMegaBytes,
								Value:      "123456",
								Timestamp:  333333,
							},
							{
								AppId:      appId,
								MetricType: models.MetricNameMemoryUsed,
								Unit:       models.UnitMegaBytes,
								Value:      "123456",
								Timestamp:  444444,
							},
						},
					}
					checkAggregatedMetricResult(components.Ports[GolangAPIServer], pathVariables, parameters, result)

					By("get the 2nd page")
					parameters = map[string]string{"start-time": "111111", "end-time": "999999", "order-direction": "asc", "page": "2", "results-per-page": "2"}
					result = AppAggregatedMetricResult{
						TotalResults: 5,
						TotalPages:   3,
						Page:         2,
						PrevUrl:      getAppAggregatedMetricUrl(appId, metricType, parameters, 1),
						NextUrl:      getAppAggregatedMetricUrl(appId, metricType, parameters, 3),
						Resources: []models.AppMetric{
							{
								AppId:      appId,
								MetricType: models.MetricNameMemoryUsed,
								Unit:       models.UnitMegaBytes,
								Value:      "123456",
								Timestamp:  555555,
							},
							{
								AppId:      appId,
								MetricType: models.MetricNameMemoryUsed,
								Unit:       models.UnitMegaBytes,
								Value:      "123456",
								Timestamp:  555556,
							},
						},
					}
					checkAggregatedMetricResult(components.Ports[GolangAPIServer], pathVariables, parameters, result)

					By("get the 3rd page")
					parameters = map[string]string{"start-time": "111111", "end-time": "999999", "order-direction": "asc", "page": "3", "results-per-page": "2"}
					result = AppAggregatedMetricResult{
						TotalResults: 5,
						TotalPages:   3,
						Page:         3,
						PrevUrl:      getAppAggregatedMetricUrl(appId, metricType, parameters, 2),
						Resources: []models.AppMetric{
							{
								AppId:      appId,
								MetricType: models.MetricNameMemoryUsed,
								Unit:       models.UnitMegaBytes,
								Value:      "123456",
								Timestamp:  666666,
							},
						},
					}
					checkAggregatedMetricResult(components.Ports[GolangAPIServer], pathVariables, parameters, result)

					By("the 4th page should be empty")
					parameters = map[string]string{"start-time": "111111", "end-time": "999999", "order-direction": "asc", "page": "4", "results-per-page": "2"}
					result = AppAggregatedMetricResult{
						TotalResults: 5,
						TotalPages:   3,
						Page:         4,
						PrevUrl:      getAppAggregatedMetricUrl(appId, metricType, parameters, 3),
						Resources:    []models.AppMetric{},
					}
					checkAggregatedMetricResult(components.Ports[GolangAPIServer], pathVariables, parameters, result)
				})
				It("should get the metrics in specified time scope", func() {
					By("get the results from 555555")
					parameters = map[string]string{"start-time": "555555", "order-direction": "asc", "page": "1", "results-per-page": "10"}
					result := AppAggregatedMetricResult{
						TotalResults: 3,
						TotalPages:   1,
						Page:         1,
						Resources: []models.AppMetric{
							{
								AppId:      appId,
								MetricType: models.MetricNameMemoryUsed,
								Unit:       models.UnitMegaBytes,
								Value:      "123456",
								Timestamp:  555555,
							},
							{
								AppId:      appId,
								MetricType: models.MetricNameMemoryUsed,
								Unit:       models.UnitMegaBytes,
								Value:      "123456",
								Timestamp:  555556,
							},
							{
								AppId:      appId,
								MetricType: models.MetricNameMemoryUsed,
								Unit:       models.UnitMegaBytes,
								Value:      "123456",
								Timestamp:  666666,
							},
						},
					}
					checkAggregatedMetricResult(components.Ports[GolangAPIServer], pathVariables, parameters, result)

					By("get the results to 444444")
					parameters = map[string]string{"end-time": "444444", "order-direction": "asc", "page": "1", "results-per-page": "10"}
					result = AppAggregatedMetricResult{
						TotalResults: 2,
						TotalPages:   1,
						Page:         1,
						Resources: []models.AppMetric{
							{
								AppId:      appId,
								MetricType: models.MetricNameMemoryUsed,
								Unit:       models.UnitMegaBytes,
								Value:      "123456",
								Timestamp:  333333,
							},
							{
								AppId:      appId,
								MetricType: models.MetricNameMemoryUsed,
								Unit:       models.UnitMegaBytes,
								Value:      "123456",
								Timestamp:  444444,
							},
						},
					}
					checkAggregatedMetricResult(components.Ports[GolangAPIServer], pathVariables, parameters, result)

					By("get the results from 444444 to 555556")
					parameters = map[string]string{"start-time": "444444", "end-time": "555556", "order-direction": "asc", "page": "1", "results-per-page": "10"}
					result = AppAggregatedMetricResult{
						TotalResults: 3,
						TotalPages:   1,
						Page:         1,
						Resources: []models.AppMetric{
							{
								AppId:      appId,
								MetricType: models.MetricNameMemoryUsed,
								Unit:       models.UnitMegaBytes,
								Value:      "123456",
								Timestamp:  444444,
							},
							{
								AppId:      appId,
								MetricType: models.MetricNameMemoryUsed,
								Unit:       models.UnitMegaBytes,
								Value:      "123456",
								Timestamp:  555555,
							},
							{
								AppId:      appId,
								MetricType: models.MetricNameMemoryUsed,
								Unit:       models.UnitMegaBytes,
								Value:      "123456",
								Timestamp:  555556,
							},
						},
					}
					checkAggregatedMetricResult(components.Ports[GolangAPIServer], pathVariables, parameters, result)
				})
			})
		})
	})
})
