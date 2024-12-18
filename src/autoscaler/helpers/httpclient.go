package helpers

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/cf"
	"code.cloudfoundry.org/lager/v3"
	"github.com/hashicorp/go-retryablehttp"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"

	"code.cloudfoundry.org/cfhttp/v2"
)

type TLSReloadTransport struct {
	Base     http.RoundTripper
	logger   lager.Logger
	tlsCerts *models.TLSCerts
}

func (t *TLSReloadTransport) tlsClientConfig() *tls.Config {
	return t.Base.(*retryablehttp.RoundTripper).Client.HTTPClient.Transport.(*http.Transport).TLSClientConfig
}

func (t *TLSReloadTransport) setTLSClientConfig(tlsConfig *tls.Config) {
	t.Base.(*retryablehttp.RoundTripper).Client.HTTPClient.Transport.(*http.Transport).TLSClientConfig = tlsConfig
}

func (t *TLSReloadTransport) certificateExpiringWithin(dur time.Duration) bool {
	x509Cert, err := x509.ParseCertificate(t.tlsClientConfig().Certificates[0].Certificate[0])
	if err != nil {
		return false
	}

	return x509Cert.NotAfter.Sub(time.Now()) < dur
}
func (t *TLSReloadTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.certificateExpiringWithin(time.Hour) {
		t.logger.Info("reloading-cert", lager.Data{"request": req})
		tlsConfig, _ := t.tlsCerts.CreateClientConfig()
		t.setTLSClientConfig(tlsConfig)
	} else {
		t.logger.Info("cert-not-expiring", lager.Data{"request": req})
	}

	return t.Base.RoundTrip(req)
}

func DefaultClientConfig() cf.ClientConfig {
	return cf.ClientConfig{
		MaxIdleConnsPerHost:     200,
		IdleConnectionTimeoutMs: 5 * 1000,
	}
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

	retryClient := cf.RetryClient(config, client, logger)

	retryClient.Transport = &TLSReloadTransport{
		Base:     retryClient.Transport,
		logger:   logger,
		tlsCerts: tlsCerts,
	}

	return retryClient, nil
}
