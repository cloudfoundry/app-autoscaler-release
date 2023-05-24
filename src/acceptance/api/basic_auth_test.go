package api_test

import (
	"fmt"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AutoScaler Health Endpoints with Basic Auth", func() {

	urlfor := func(name string, path string) func() string {
		return func() string {
			if path != "" {
				healthURL = fmt.Sprintf("%s%s", cfg.ASApiEndpoint, path)
			}
			healthURL := strings.Replace(healthURL, cfg.ServiceName, cfg.ServiceName+"-"+name, 1)
			return healthURL
		}
	}
	DescribeTable("Basic Auth Credentials not provided",
		func(url func() string, statusCode func() int) {
			Expect(Get(url())).To(Equal(statusCode()), "to get status code %d when getting %s", statusCode(), url())
		},
		Entry("API Server", urlfor("apiserver", ""), getStatus),
		Entry("Eventgenerator", urlfor("eventgenerator", ""), getStatus),
		Entry("Scaling Engine", urlfor("scalingengine", ""), getStatus),
		Entry("Operator", urlfor("operator", ""), getStatus),
		Entry("Metrics Forwarder", urlfor("metricsforwarder", ""), getStatus),
		Entry("Scheduler", urlfor("scheduler", ""), getStatus),
	)

	DescribeTable("Basic Auth Credentials Provided",

		func(url func() string, statusCode func() int) {
			cfg.HealthEndpointsBasicAuthEnabled = true
			Expect(Get(url())).To(Equal(statusCode()), "to get status code %d when getting %s", statusCode(), url())
		},
		Entry("API Server", urlfor("apiserver", ""), getStatus),
		Entry("Eventgenerator", urlfor("eventgenerator", ""), getStatus),
		Entry("Scaling Engine", urlfor("scalingengine", ""), getStatus),
		Entry("Operator", urlfor("operator", ""), getStatus),
		Entry("Metrics Forwarder", urlfor("metricsforwarder", ""), getStatus),
		Entry("Scheduler", urlfor("scheduler", ""), getStatus),
	)

	DescribeTable("Liveness with Basic Auth Credentials Provided",

		func(url func() string, statusCode func() int) {
			cfg.HealthEndpointsBasicAuthEnabled = true
			Expect(Get(url())).To(Equal(statusCode()), "to get status code %d when getting %s", statusCode(), url())
		},
		Entry("API Server", urlfor("apiserver", "/health"), getStatus),
		Entry("Eventgenerator", urlfor("eventgenerator", "/health"), getStatus),
		Entry("Scaling Engine", urlfor("scalingengine", "/health"), getStatus),
		Entry("Operator", urlfor("operator", "/health"), getStatus),
		Entry("Metrics Forwarder", urlfor("metricsforwarder", "/health"), getStatus),
		Entry("Scheduler", urlfor("scheduler", "/health"), getStatus),
	)
})

func getStatus() int {
	if cfg.HealthEndpointsBasicAuthEnabled {
		return 401
	} else {
		return 200
	}
}

func Get(url string) int {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	Expect(err).ShouldNot(HaveOccurred())
	resp, err := client.Do(req)
	Expect(err).ShouldNot(HaveOccurred())
	defer func() { _ = resp.Body.Close() }()
	Expect(err).ShouldNot(HaveOccurred())
	return resp.StatusCode
}
