package app

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strconv"

	api "code.cloudfoundry.org/app-autoscaler-release/src/acceptance/assets/app/go_app/internal/custommetrics"
	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	json "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
)

//counterfeiter:generate . CustomMetricClient
type CustomMetricClient interface {
	PostCustomMetric(ctx context.Context, appConfig *cfenv.App, metricsValue float64, metricName string, useMtls bool) error
}

type CustomMetricAPIClient struct{}

var _ CustomMetricClient = &CustomMetricAPIClient{}

func CustomMetricsTests(logger logr.Logger, r *gin.RouterGroup, customMetricTest CustomMetricClient) *gin.RouterGroup {

	r.GET("/mtls/:name/:value", handleCustomMetricsEndpoint(customMetricTest, true))
	r.GET("/:name/:value", handleCustomMetricsEndpoint(customMetricTest, false))

	return r
}

func handleCustomMetricsEndpoint(customMetricTest CustomMetricClient, useMtls bool) func(c *gin.Context) {
	return func(c *gin.Context) {
		var (
			metricName  string
			metricValue uint64
		)
		var err error
		if metricName = c.Param("name"); metricName == "" {
			Error(c, http.StatusBadRequest, "empty metric name")
			return
		}
		if metricValue, err = strconv.ParseUint(c.Param("value"), 10, 64); err != nil {
			Error(c, http.StatusBadRequest, "invalid metric value: %s", err.Error())
			return
		}

		err = customMetricTest.PostCustomMetric(c, nil, float64(metricValue), metricName, useMtls)
		if err != nil {
			Error(c, http.StatusInternalServerError, "failed to submit custom metric: %s", err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{"mtls": useMtls})
	}
}

func (_ *CustomMetricAPIClient) PostCustomMetric(ctx context.Context, appConfig *cfenv.App, metricValue float64, metricName string, useMtls bool) error {
	var err error
	if appConfig == nil {
		appConfig, err = cfenv.Current()
		if err != nil {
			return fmt.Errorf("cloud foundry environment not found %w", err)

		}
	}

	appId := appConfig.AppID
	autoscalerCredentials, err := getAutoscalerCredentials(appConfig)
	if err != nil {
		return err
	}

	var (
		autoscalerApiServerURL string
		client                 *http.Client
	)
	if useMtls {
		autoscalerApiServerURL = autoscalerCredentials.MtlsURL
		if autoscalerApiServerURL == "" {
			// fallback for the "build in" case
			autoscalerApiServerURL = autoscalerCredentials.URL
		}
		client, err = getCFInstanceIdentityCertificateClient()
	} else {
		autoscalerApiServerURL = autoscalerCredentials.URL
		client, err = getBasicAuthClient(autoscalerCredentials)
	}
	if err != nil {
		return err
	}

	apiClient, err := api.NewClient(autoscalerApiServerURL, api.WithClient(client))
	if err != nil {
		return err
	}

	metrics := createSingletonMetric(metricName, metricValue)

	params := api.V1AppsAppGuidMetricsPostParams{AppGuid: appId}

	return apiClient.V1AppsAppGuidMetricsPost(ctx, metrics, params)
}

func getBasicAuthClient(autoscalerCredentials api.CustomMetricsCredentials) (*http.Client, error) {
	return api.NewBasicAuthTransport(autoscalerCredentials).Client(), nil
}

func getAutoscalerCredentials(appConfig *cfenv.App) (api.CustomMetricsCredentials, error) {
	result := api.CustomMetricsCredentials{}

	customMetricEnv := os.Getenv("AUTO_SCALER_CUSTOM_METRIC_ENV")

	if customMetricEnv != "" {
		err := json.Unmarshal([]byte(customMetricEnv), &result)
		if err != nil {
			return result, err
		}
	} else {
		autoscalers, err := appConfig.Services.WithTag("app-autoscaler")
		if err != nil {
			return result, err
		}
		autoscalerCredentials := autoscalers[0].Credentials["custom_metrics"]

		err = mapstructure.Decode(autoscalerCredentials, &result)

		if err != nil {
			return result, err
		}
	}

	return result, nil
}

func createSingletonMetric(metricName string, metricValue float64) *api.Metrics {
	metric := api.Metric{
		Name:  metricName,
		Value: metricValue,
	}

	metrics := &api.Metrics{
		InstanceIndex: 0,
		Metrics:       []api.Metric{metric},
	}
	return metrics
}

func getCFInstanceIdentityCertificateClient() (*http.Client, error) {
	cfInstanceKey := os.Getenv("CF_INSTANCE_KEY")
	cfInstanceCert := os.Getenv("CF_INSTANCE_CERT")

	cert, err := tls.LoadX509KeyPair(cfInstanceCert, cfInstanceKey)
	if err != nil {
		return nil, fmt.Errorf("error loading CF instance identity x509 keypair: %w", err)
	}

	caCertBytes, err := os.ReadFile(cfInstanceCert)
	if err != nil {
		return nil, fmt.Errorf("error reading CF CA: %w", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertBytes)

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		RootCAs:            caCertPool,
	}

	t := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client := &http.Client{Transport: t}
	return client, nil
}
