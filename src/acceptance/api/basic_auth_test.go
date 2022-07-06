package api_test

import (
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AutoScaler Basic Auth Tests", func() {

	Context("API Server: basic auth tests", func() {
		It("should succeed to check health", func() {
			req, err := http.NewRequest("GET", healthURL, nil)
			Expect(err).ShouldNot(HaveOccurred())
			resp, err := DoAPIRequest(req)
			Expect(err).ShouldNot(HaveOccurred())

			defer func() { _ = resp.Body.Close() }()

			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))
		})
	})

	Context("Eventgenerator: basic auth tests", func() {
		It("should fail to check health without basic auth credentials", func() {
			req, err := http.NewRequest("GET", strings.Replace(healthURL, cfg.ServiceName, cfg.ServiceName+"-eventgenerator", 1), nil)
			Expect(err).ShouldNot(HaveOccurred())
			resp, err := DoAPIRequest(req)
			Expect(err).ShouldNot(HaveOccurred())

			defer func() { _ = resp.Body.Close() }()
			Expect(err).ShouldNot(HaveOccurred())
			if cfg.HealthEndpointsBasicAuthEnabled {
				Expect(resp.StatusCode).To(Equal(401))
			} else {
				Expect(resp.StatusCode).To(Equal(200))
			}
		})
	})

	Context("Scaling Engine: basic auth tests", func() {
		It("should fail to check health without basic auth credentials", func() {
			req, err := http.NewRequest("GET", strings.Replace(healthURL, cfg.ServiceName, cfg.ServiceName+"-scalingengine", 1), nil)
			Expect(err).ShouldNot(HaveOccurred())
			resp, err := DoAPIRequest(req)
			Expect(err).ShouldNot(HaveOccurred())

			defer func() { _ = resp.Body.Close() }()

			Expect(err).ShouldNot(HaveOccurred())
			if cfg.HealthEndpointsBasicAuthEnabled {
				Expect(resp.StatusCode).To(Equal(401))
			} else {
				Expect(resp.StatusCode).To(Equal(200))
			}
		})
	})

	Context("Operator: basic auth tests", func() {
		It("should fail to check health without basic auth credentials", func() {
			req, err := http.NewRequest("GET", strings.Replace(healthURL, cfg.ServiceName, cfg.ServiceName+"-operator", 1), nil)
			Expect(err).ShouldNot(HaveOccurred())
			resp, err := DoAPIRequest(req)
			Expect(err).ShouldNot(HaveOccurred())

			defer func() { _ = resp.Body.Close() }()

			Expect(err).ShouldNot(HaveOccurred())
			if cfg.HealthEndpointsBasicAuthEnabled {
				Expect(resp.StatusCode).To(Equal(401))
			} else {
				Expect(resp.StatusCode).To(Equal(200))
			}
		})
	})

	Context("Metrics Server: basic auth tests", func() {
		It("should fail to check health without basic auth credentials", func() {
			req, err := http.NewRequest("GET", strings.Replace(healthURL, cfg.ServiceName, cfg.ServiceName+"-metricsserver", 1), nil)
			Expect(err).ShouldNot(HaveOccurred())
			resp, err := DoAPIRequest(req)
			Expect(err).ShouldNot(HaveOccurred())

			defer func() { _ = resp.Body.Close() }()

			Expect(err).ShouldNot(HaveOccurred())
			if cfg.HealthEndpointsBasicAuthEnabled {
				Expect(resp.StatusCode).To(Equal(401))
			} else {
				Expect(resp.StatusCode).To(Equal(200))
			}
		})
	})

	Context("Metrics Gateway: basic auth tests", func() {
		It("should fail to check health without basic auth credentials", func() {
			req, err := http.NewRequest("GET", strings.Replace(healthURL, cfg.ServiceName, cfg.ServiceName+"-metricsgateway", 1), nil)
			Expect(err).ShouldNot(HaveOccurred())
			resp, err := DoAPIRequest(req)
			Expect(err).ShouldNot(HaveOccurred())

			defer func() { _ = resp.Body.Close() }()

			Expect(err).ShouldNot(HaveOccurred())
			if cfg.HealthEndpointsBasicAuthEnabled {
				Expect(resp.StatusCode).To(Equal(401))
			} else {
				Expect(resp.StatusCode).To(Equal(200))
			}
		})
	})

	Context("Metrics Forwarder: basic auth tests", func() {
		It("should fail to check health without basic auth credentials", func() {
			req, err := http.NewRequest("GET", strings.Replace(healthURL, cfg.ServiceName, cfg.ServiceName+"-metricsforwarder", 1), nil)
			Expect(err).ShouldNot(HaveOccurred())
			resp, err := DoAPIRequest(req)
			Expect(err).ShouldNot(HaveOccurred())

			defer func() { _ = resp.Body.Close() }()

			Expect(err).ShouldNot(HaveOccurred())
			if cfg.HealthEndpointsBasicAuthEnabled {
				Expect(resp.StatusCode).To(Equal(401))
			} else {
				Expect(resp.StatusCode).To(Equal(200))
			}
		})
	})

	Context("Scheduler: basic auth tests", func() {
		It("should fail to check health without basic auth credentials", func() {
			req, err := http.NewRequest("GET", strings.Replace(healthURL, cfg.ServiceName, cfg.ServiceName+"-scheduler", 1), nil)
			Expect(err).ShouldNot(HaveOccurred())
			resp, err := DoAPIRequest(req)
			Expect(err).ShouldNot(HaveOccurred())

			defer func() { _ = resp.Body.Close() }()

			Expect(err).ShouldNot(HaveOccurred())
			if cfg.HealthEndpointsBasicAuthEnabled {
				Expect(resp.StatusCode).To(Equal(401))
			} else {
				Expect(resp.StatusCode).To(Equal(200))
			}
		})
	})

})
