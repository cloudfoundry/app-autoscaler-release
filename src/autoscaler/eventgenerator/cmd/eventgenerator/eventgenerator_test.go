package main_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/eventgenerator/config"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/testhelpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Eventgenerator", func() {
	var (
		runner      *EventGeneratorRunner
		httpsClient *http.Client
		serverURL   string
	)

	BeforeEach(func() {
		runner = NewEventGeneratorRunner()
		httpsClient = testhelpers.NewEventGeneratorClient()
		serverURL = fmt.Sprintf("https://127.0.0.1:%d", conf.Server.Port)
	})

	AfterEach(func() {
		runner.KillWithFire()
	})

	Context("with a valid config file", func() {
		BeforeEach(func() {
			runner.Start()
		})

		It("Starts successfully, retrives metrics and  generates events", func() {
			Consistently(runner.Session).ShouldNot(Exit())
			Eventually(func() bool { return mockLogCache.ReadRequestsCount() >= 1 }, 5*time.Second).Should(BeTrue())
			Eventually(func() bool { return len(mockScalingEngine.ReceivedRequests()) >= 1 }, time.Duration(2*breachDurationSecs)*time.Second).Should(BeTrue())
		})
	})

	Context("with a missing config file", func() {
		BeforeEach(func() {
			runner.startCheck = ""
			runner.configPath = "bogus"
			runner.Start()
		})

		It("fails with an error", func() {
			Eventually(runner.Session).Should(Exit(1))
			Expect(runner.Session.Buffer()).To(Say("failed to open config file"))
		})
	})

	Context("with an invalid config file", func() {
		BeforeEach(func() {
			runner.startCheck = ""
			badfile, err := os.CreateTemp("", "bad-mc-config")
			Expect(err).NotTo(HaveOccurred())
			runner.configPath = badfile.Name()
			// #nosec G306
			err = os.WriteFile(runner.configPath, []byte("bogus"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
			runner.Start()
		})

		AfterEach(func() {
			os.Remove(runner.configPath)
		})

		It("fails with an error", func() {
			Eventually(runner.Session).Should(Exit(1))
			Expect(runner.Session.Buffer()).To(Say("failed to parse config file"))
		})
	})

	Context("with missing configuration", func() {
		BeforeEach(func() {
			runner.startCheck = ""
			conf := &config.Config{
				Logging: helpers.LoggingConfig{
					Level: "debug",
				},
				Aggregator: config.AggregatorConfig{
					AggregatorExecuteInterval: 2 * time.Second,
					PolicyPollerInterval:      2 * time.Second,
					MetricPollerCount:         2,
					AppMonitorChannelSize:     2,
				},
				Evaluator: config.EvaluatorConfig{
					EvaluationManagerInterval: 2 * time.Second,
					EvaluatorCount:            2,
					TriggerArrayChannelSize:   2,
				},
			}
			configFile := writeConfig(conf)
			runner.configPath = configFile.Name()
			runner.Start()
		})

		AfterEach(func() {
			os.Remove(runner.configPath)
		})

		It("should fail validation", func() {
			Eventually(runner.Session).Should(Exit(1))
			Expect(runner.Session.Buffer()).To(Say("failed to validate configuration"))
		})
	})

	Context("when an interrupt is sent", func() {
		BeforeEach(func() {
			runner.Start()
		})

		It("should stop", func() {
			runner.Session.Interrupt()
			Eventually(runner.Session, 5).Should(Exit(0))
		})
	})

	Describe("EventGenerator REST API", func() {
		Context("when a request for aggregated metrics history comes", func() {
			BeforeEach(func() {
				runner.Start()
			})

			It("returns with a 200", func() {
				rsp, err := httpClient.Get(fmt.Sprintf("%s/v1/apps/an-app-id/aggregated_metric_histories/a-metric-type", serverURL))
				Expect(err).NotTo(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusOK))
				rsp.Body.Close()
			})

		})

	})

	Describe("when Health server is ready to serve RESTful API", func() {
		BeforeEach(func() {
			basicAuthConfig := conf
			basicAuthConfig.Health.HealthCheckUsername = ""
			basicAuthConfig.Health.HealthCheckPassword = ""
			runner.configPath = writeConfig(&basicAuthConfig).Name()

			runner.Start()

		})

		Context("when a request to query health comes", func() {
			It("returns with a 200", func() {
				rsp, err := httpsClient.Get(fmt.Sprintf("%s/health", serverURL))
				Expect(err).NotTo(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusOK))

				raw, err := io.ReadAll(rsp.Body)
				Expect(err).NotTo(HaveOccurred())

				healthData := string(raw)
				Expect(healthData).To(ContainSubstring("autoscaler_eventgenerator_concurrent_http_request"))
				Expect(healthData).To(ContainSubstring("autoscaler_eventgenerator_policyDB"))
				Expect(healthData).To(ContainSubstring("autoscaler_eventgenerator_appMetricDB"))
				Expect(healthData).To(ContainSubstring("go_goroutines"))
				Expect(healthData).To(ContainSubstring("go_memstats_alloc_bytes"))
				rsp.Body.Close()

			})
		})
	})

	Describe("when Health server is ready to serve RESTful API with basic Auth", func() {
		BeforeEach(func() {
			runner.Start()
		})
		Context("when username and password are incorrect for basic authentication during health check", func() {
			It("should return 401", func() {

				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/health", serverURL), nil)
				Expect(err).NotTo(HaveOccurred())

				req.SetBasicAuth("wrongusername", "wrongpassword")

				rsp, err := httpsClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when username and password are correct for basic authentication during health check", func() {
			It("should return 200", func() {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/health", serverURL), nil)
				Expect(err).NotTo(HaveOccurred())

				req.SetBasicAuth(conf.Health.HealthCheckUsername, conf.Health.HealthCheckPassword)

				rsp, err := httpsClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("when Health server is ready to serve RESTful API with basic Auth", func() {
		BeforeEach(func() {
			runner.Start()
		})
		Context("when username and password are incorrect for basic authentication during health check", func() {
			It("should return 401", func() {

				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/health", serverURL), nil)
				Expect(err).NotTo(HaveOccurred())

				req.SetBasicAuth("wrongusername", "wrongpassword")

				rsp, err := httpsClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when username and password are correct for basic authentication during health check", func() {
			It("should return 200", func() {
				// Load the client key and certificate
				//
				// Load your custom certificate file

				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/health", serverURL), nil)
				Expect(err).NotTo(HaveOccurred())

				req.SetBasicAuth(conf.Health.HealthCheckUsername, conf.Health.HealthCheckPassword)

				rsp, err := httpsClient.Do(req)
				Expect(err).ToNot(HaveOccurred())

				body, err := io.ReadAll(rsp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(body)).To(ContainSubstring("autoscaler_eventgenerator_concurrent_http_request"))

				Expect(rsp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})
})
