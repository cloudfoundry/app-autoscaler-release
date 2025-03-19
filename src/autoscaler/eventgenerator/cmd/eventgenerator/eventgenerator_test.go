package main_test

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/configutil"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/eventgenerator/config"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/testhelpers"
	. "code.cloudfoundry.org/app-autoscaler/src/autoscaler/testhelpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Eventgenerator", func() {
	var (
		runner                      *EventGeneratorRunner
		httpClientForEventGenerator *http.Client
		httpClientForHealth         *http.Client

		serverURL   *url.URL
		healthURL   *url.URL
		cfServerURL *url.URL

		vcapPort int
		err      error
	)

	BeforeEach(func() {
		runner = NewEventGeneratorRunner()

		vcapPort = 8090 + GinkgoParallelProcess()

		httpClientForEventGenerator = NewEventGeneratorClient()
		httpClientForHealth = &http.Client{}

		serverURL, err = url.Parse("https://127.0.0.1:" + strconv.Itoa(conf.Server.Port))
		Expect(err).ToNot(HaveOccurred())

		healthURL, err = url.Parse("http://127.0.0.1:" + strconv.Itoa(conf.Health.ServerConfig.Port))
		Expect(err).ToNot(HaveOccurred())

		cfServerURL, err = url.Parse("http://127.0.0.1:" + strconv.Itoa(conf.CFServer.Port))
		Expect(err).ToNot(HaveOccurred())

	})

	JustBeforeEach(func() {
		runner.Start()
	})

	AfterEach(func() {
		runner.KillWithFire()
	})
	Context("with a valid config file", func() {
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
		})

		AfterEach(func() {
			os.Remove(runner.configPath)
		})

		It("fails with an error", func() {
			Eventually(runner.Session).Should(Exit(1))
			Expect(runner.Session.Buffer()).To(Say("failed to read config file"))
		})
	})

	Context("with missing configuration", func() {
		BeforeEach(func() {
			runner.startCheck = ""
			conf := &config.Config{
				Logging: helpers.LoggingConfig{
					Level: "debug",
				},
				Aggregator: &config.AggregatorConfig{
					AggregatorExecuteInterval: 2 * time.Second,
					PolicyPollerInterval:      2 * time.Second,
					MetricPollerCount:         2,
					AppMonitorChannelSize:     2,
				},
				Evaluator: &config.EvaluatorConfig{
					EvaluationManagerInterval: 2 * time.Second,
					EvaluatorCount:            2,
					TriggerArrayChannelSize:   2,
				},
			}
			configFile := writeConfig(conf)
			runner.configPath = configFile.Name()
		})

		AfterEach(func() {
			os.Remove(runner.configPath)
		})

		It("should fail validation", func() {
			Eventually(runner.Session).Should(Exit(1))
			Expect(runner.Session.Buffer()).To(Say("failed to validate configuration"))
		})
	})

	When("an interrupt is sent", func() {

		It("should stop", func() {
			runner.Session.Interrupt()
			Eventually(runner.Session, 5).Should(Exit(0))
		})
	})

	Describe("EventGenerator REST API", func() {
		When("a request for aggregated metrics history comes", func() {
			BeforeEach(func() {
				serverURL.Path = "/v1/apps/an-app-id/aggregated_metric_histories/a-metric-type"
			})

			It("returns with a 200", func() {
				rsp, err := httpClientForEventGenerator.Get(serverURL.String())
				Expect(err).NotTo(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusOK))
				rsp.Body.Close()
			})
		})
	})

	Describe("EventGenerator Health endpoint", func() {

		BeforeEach(func() {
			serverURL.Path = "/health"
		})

		When("Health server is ready to serve RESTful API", func() {
			BeforeEach(func() {
				basicAuthConfig := conf
				basicAuthConfig.Health.BasicAuth.Username = ""
				basicAuthConfig.Health.BasicAuth.Password = ""
				runner.configPath = writeConfig(&basicAuthConfig).Name()

			})

			When("a request to query health comes", func() {
				It("returns with a 200", func() {
					rsp, err := httpClientForHealth.Get(healthURL.String())
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

		When("Health server is ready to serve RESTful API with basic Auth", func() {

			When("username and password are incorrect for basic authentication during health check", func() {
				It("should return 401", func() {
					req, err := http.NewRequest(http.MethodGet, healthURL.String(), nil)
					Expect(err).NotTo(HaveOccurred())

					req.SetBasicAuth("wrongusername", "wrongpassword")

					rsp, err := httpClientForHealth.Do(req)
					Expect(err).ToNot(HaveOccurred())
					Expect(rsp.StatusCode).To(Equal(http.StatusUnauthorized))
				})
			})

			When("username and password are correct for basic authentication during health check", func() {
				It("should return 200", func() {
					req, err := http.NewRequest(http.MethodGet, healthURL.String(), nil)
					Expect(err).NotTo(HaveOccurred())

					req.SetBasicAuth(conf.Health.BasicAuth.Username, conf.Health.BasicAuth.Password)

					rsp, err := httpClientForHealth.Do(req)
					Expect(err).ToNot(HaveOccurred())
					Expect(rsp.StatusCode).To(Equal(http.StatusOK))
				})
			})
		})

		When("Health server is ready to serve RESTful API with basic Auth", func() {

			When("username and password are incorrect for basic authentication during health check", func() {
				It("should return 401", func() {
					req, err := http.NewRequest(http.MethodGet, healthURL.String(), nil)
					Expect(err).NotTo(HaveOccurred())

					req.SetBasicAuth("wrongusername", "wrongpassword")

					rsp, err := httpClientForHealth.Do(req)
					Expect(err).ToNot(HaveOccurred())
					Expect(rsp.StatusCode).To(Equal(http.StatusUnauthorized))
				})
			})

			When("username and password are correct for basic authentication during health check", func() {
				It("should return 200", func() {
					req, err := http.NewRequest(http.MethodGet, healthURL.String(), nil)
					Expect(err).NotTo(HaveOccurred())

					req.SetBasicAuth(conf.Health.BasicAuth.Username, conf.Health.BasicAuth.Password)

					rsp, err := httpClientForHealth.Do(req)
					Expect(err).ToNot(HaveOccurred())
					Expect(rsp.StatusCode).To(Equal(http.StatusOK))
				})
			})
		})
	})

	When("running CF server", func() {
		Context("Get /v1/liveness", func() {
			It("should return 200", func() {
				cfServerURL.Path = "/v1/liveness"

				req, err := http.NewRequest(http.MethodGet, cfServerURL.String(), nil)
				Expect(err).NotTo(HaveOccurred())

				err = SetXFCCCertHeader(req, conf.CFServer.XFCC.ValidOrgGuid, conf.CFServer.XFCC.ValidSpaceGuid)
				Expect(err).NotTo(HaveOccurred())

				rsp, err := healthHttpClient.Do(req)
				Expect(err).ToNot(HaveOccurred())

				Expect(rsp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})

	When("running CF server", func() {
		var (
			cfInstanceKeyFile  string
			cfInstanceCertFile string
		)

		BeforeEach(func() {
			runner = NewEventGeneratorCFRunner()
			rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			Expect(err).NotTo(HaveOccurred())

			cfInstanceCert, err := testhelpers.GenerateClientCertWithPrivateKey("org-guid", "space-guid", rsaPrivateKey)
			Expect(err).NotTo(HaveOccurred())

			certTmpDir := os.TempDir()

			cfInstanceCertFile, err := configutil.MaterializeContentInFile(certTmpDir, "eventgenerator.crt", string(cfInstanceCert))
			Expect(err).NotTo(HaveOccurred())
			os.Setenv("CF_INSTANCE_CERT", cfInstanceCertFile)

			cfInstanceKey := testhelpers.GenerateClientKeyWithPrivateKey(rsaPrivateKey)
			cfInstanceKeyFile, err = configutil.MaterializeContentInFile(certTmpDir, "eventgenerator.key", string(cfInstanceKey))
			Expect(err).NotTo(HaveOccurred())
			os.Setenv("CF_INSTANCE_KEY", cfInstanceKeyFile)

			os.Setenv("VCAP_APPLICATION", "{}")
			conf.Db = config.DbConfig{}
			conf.Evaluator = nil
			conf.Aggregator = nil
			conf.CircuitBreaker = nil
			conf.HttpClientTimeout = nil
			os.Setenv("VCAP_SERVICES", getVcapServices(conf))
			os.Setenv("PORT", fmt.Sprintf("%d", vcapPort))
		})

		AfterEach(func() {
			runner.Interrupt()
			Eventually(runner.Session, 5).Should(Exit(0))

			os.Remove(cfInstanceKeyFile)
			os.Remove(cfInstanceCertFile)

			os.Unsetenv("CF_INSTANCE_KEY")
			os.Unsetenv("CF_INSTANCE_CERT")
			os.Unsetenv("VCAP_APPLICATION")
			os.Unsetenv("VCAP_SERVICES")
			os.Unsetenv("PORT")
		})

		It("Starts successfully, retrives metrics and generates events", func() {
			Consistently(runner.Session).ShouldNot(Exit())
			// refactor mocklogcache to restart on before each to avoid shared resource
			// Eventually(func() bool { return mockLogCache.ReadRequestsCount() >= 1 }, 5*time.Second).Should(BeTrue())
			// Eventually(func() bool { return len(mockScalingEngine.ReceivedRequests()) >= 1 }, time.Duration(2*breachDurationSecs)*time.Second).Should(BeTrue())
		})
	})
})

func getVcapServices(conf config.Config) (result string) {
	var dbType string

	configJson, err := conf.ToJSON()
	Expect(err).NotTo(HaveOccurred())

	dbClientCert, err := os.ReadFile("../../../../../test-certs/postgres.crt")
	Expect(err).NotTo(HaveOccurred())
	dbClientKey, err := os.ReadFile("../../../../../test-certs/postgres.key")
	Expect(err).NotTo(HaveOccurred())
	dbClientCA, err := os.ReadFile("../../../../../test-certs/autoscaler-ca.crt")
	Expect(err).NotTo(HaveOccurred())

	dbURL := os.Getenv("DBURL")
	Expect(dbURL).NotTo(BeEmpty())
	if strings.Contains(dbURL, "postgres") {
		dbType = "postgres"
	} else {
		dbType = "mysql"
	}

	result = `{
			"user-provided": [
			  { "name": "eventgenerator-config", "tags": ["eventgenerator-config"], "credentials": { "eventgenerator-config": ` + configJson + `}}
            ],
			"autoscaler": [ {
				"name": "some-service",
				"credentials": {
					"uri": "` + dbURL + `",
					"client_cert": "` + strings.ReplaceAll(string(dbClientCert), "\n", "\\n") + `",
					"client_key": "` + strings.ReplaceAll(string(dbClientKey), "\n", "\\n") + `",
					"server_ca": "` + strings.ReplaceAll(string(dbClientCA), "\n", "\\n") + `"
				},
				"syslog_drain_url": "",
				"tags": ["policy_db", "app_metrics_db", "` + dbType + `"]
			}]}` // #nosec G101

	return result
}
