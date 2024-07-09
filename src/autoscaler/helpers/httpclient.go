package helpers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/cf"
	"code.cloudfoundry.org/lager/v3"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"

	"code.cloudfoundry.org/cfhttp/v2"
)

func DefaultClientConfig() cf.ClientConfig {
	return cf.ClientConfig{
		MaxIdleConnsPerHost:     200,
		IdleConnectionTimeoutMs: 5 * 1000,
	}
}

func CreateHTTPClient(ba *models.BasicAuth, config cf.ClientConfig, logger lager.Logger) (*http.Client, error) {
	auth := ba.Username + ":" + ba.Password
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))

	client := cfhttp.NewClient(
		cfhttp.WithDialTimeout(30*time.Second),
		cfhttp.WithIdleConnTimeout(time.Duration(config.IdleConnectionTimeoutMs)*time.Millisecond),
		cfhttp.WithMaxIdleConnsPerHost(config.MaxIdleConnsPerHost),
	)

	client.Transport = &TransportWithBasicAuth{
		Username:  ba.Username,
		Password:  ba.Password,
		BasicAuth: basicAuth,
	}

	return cf.RetryClient(config, client, logger), nil
}

// TransportWithBasicAuth is a custom Transport that adds Basic Authentication headers to the request
type TransportWithBasicAuth struct {
	Username  string
	Password  string
	BasicAuth string
}

func (t *TransportWithBasicAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", t.BasicAuth)
	return http.DefaultTransport.RoundTrip(req)
}

func CreateHTTPSClient(tlsCerts *models.TLSCerts, config cf.ClientConfig, logger lager.Logger) (*http.Client, error) {
	tlsConfig, err := tlsCerts.CreateClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create tls config: %w", err)
	}

	client := cfhttp.NewClient(
		cfhttp.WithTLSConfig(tlsConfig),
		cfhttp.WithDialTimeout(30*time.Second),
		cfhttp.WithIdleConnTimeout(time.Duration(config.IdleConnectionTimeoutMs)*time.Millisecond),
		cfhttp.WithMaxIdleConnsPerHost(config.MaxIdleConnsPerHost),
	)

	return cf.RetryClient(config, client, logger), nil
}
