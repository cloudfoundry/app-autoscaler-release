package app

import "strings"

type CurlConfig interface {
	GetAppsDomain() string
	Protocol() string
	GetSkipSSLValidation() bool
}

type AppUriCreator struct {
	CurlConfig CurlConfig
}

type uriCreator interface {
	AppUri(appName, path string) string
}

func (uriCreator *AppUriCreator) AppUri(appName string, path string) string {
	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	var subdomain string
	if appName != "" {
		subdomain = appName + "."
	}

	return uriCreator.CurlConfig.Protocol() + subdomain + uriCreator.CurlConfig.GetAppsDomain() + path
}
