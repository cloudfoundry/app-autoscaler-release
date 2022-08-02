package helpers

import (
	"net"
	"net/http"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/cf"
	"code.cloudfoundry.org/lager"
	"github.com/hashicorp/go-retryablehttp"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"

	"code.cloudfoundry.org/cfhttp"
)

func CreateHTTPClient(tlsCerts *models.TLSCerts, logger lager.Logger, clientName string) (*http.Client, error) {
	if tlsCerts.CertFile == "" || tlsCerts.KeyFile == "" {
		tlsCerts = nil
	}

	//nolint:staticcheck //TODO https://github.com/cloudfoundry/app-autoscaler-release/issues/549
	client := cfhttp.NewClient()
	if tlsCerts != nil {
		//nolint:staticcheck  // SA1019 TODO: https://github.com/cloudfoundry/app-autoscaler-release/issues/548
		tlsConfig, err := cfhttp.NewTLSConfig(tlsCerts.CertFile, tlsCerts.KeyFile, tlsCerts.CACertFile)
		if err != nil {
			return nil, err
		}
		client.Transport.(*http.Transport).TLSClientConfig = tlsConfig
		client.Transport.(*http.Transport).DialContext = (&net.Dialer{
			Timeout: 30 * time.Second,
		}).DialContext
	}

	retryClient := retryablehttp.NewClient()
	retryClient.Logger = cf.LeveledLoggerAdapter{Logger: logger.Session(clientName)}
	retryClient.HTTPClient = client
	retryClient.ErrorHandler = func(resp *http.Response, err error, numTries int) (*http.Response, error) {
		return resp, err
	}
	return retryClient.StandardClient(), nil
}
