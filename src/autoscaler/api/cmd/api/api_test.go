package main_test

import (
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/routes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	. "code.cloudfoundry.org/app-autoscaler/src/autoscaler/testhelpers"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/api/config"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/db"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Api", func() {
	var (
		runner           *ApiRunner
		rsp              *http.Response
		brokerHttpClient *http.Client
	)

	BeforeEach(func() {
		brokerHttpClient = NewServiceBrokerClient()
		runner = NewApiRunner()
	})

	Describe("Api configuration check", func() {
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
				badfile, err := os.CreateTemp("", "bad-ap-config")
				Expect(err).NotTo(HaveOccurred())
				runner.configPath = badfile.Name()
				err = os.WriteFile(runner.configPath, []byte("bogus"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
				runner.Start()
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
				missingConfig := cfg

				missingConfig.DB = make(map[string]db.DatabaseConfig)
				missingConfig.DB[db.PolicyDb] = db.DatabaseConfig{URL: ""}
				missingConfig.DB[db.BindingDb] = db.DatabaseConfig{URL: ""}

				var brokerCreds []config.BrokerCredentialsConfig
				missingConfig.BrokerCredentials = brokerCreds

				missingConfig.BrokerServer.Port = 7000 + GinkgoParallelProcess()
				missingConfig.Logging.Level = "debug"
				runner.configPath = writeConfig(&missingConfig).Name()
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

	})

	Describe("when interrupt is sent", func() {
		BeforeEach(func() {
			runner.Start()
		})

		It("should stop", func() {
			runner.Session.Interrupt()
			Eventually(runner.Session, 5).Should(Exit(0))
		})

	})

	Describe("BuildIn Mode", func() {
		AfterEach(func() {
			runner.Interrupt()
			Eventually(runner.Session, 5).Should(Exit(0))
		})
		Context("BuildIn Mode is false", func() {
			BeforeEach(func() {
				basicAuthConfig := cfg
				basicAuthConfig.UseBuildInMode = false
				runner.startCheck = ""
				runner.Start()
			})
			It("should start both broker and public-api", func() {
				Eventually(runner.Session.Buffer, 2*time.Second).Should(Say("api.broker_http_server.broker-http-server-created"))
				Eventually(runner.Session.Buffer, 2*time.Second).Should(Say("api.public_api_http_server.public-api-http-server-created"))
				Eventually(runner.Session.Buffer, 2*time.Second).Should(Say("api.started"))
			})
		})

		Context("BuildIn Mode is true", func() {
			BeforeEach(func() {
				basicAuthConfig := cfg
				basicAuthConfig.UseBuildInMode = true
				runner.startCheck = ""
				runner.Start()
			})
			It("should start not start broker ", func() {
				Eventually(runner.Session.Buffer, 2*time.Second).ShouldNot(Say("api.broker_http_server.broker-http-server-created"))
				Eventually(runner.Session.Buffer, 2*time.Second).Should(Say("api.public_api_http_server.public-api-http-server-created"))
				Eventually(runner.Session.Buffer, 2*time.Second).Should(Say("api.started"))
			})
		})
	})

	Describe("Broker Rest API", func() {
		AfterEach(func() {
			runner.Interrupt()
			Eventually(runner.Session, 5).Should(Exit(0))
		})
		Context("When a request comes to broker catalog", func() {
			BeforeEach(func() {
				runner.Start()
			})
			It("succeeds with a 200", func() {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://127.0.0.1:%d/v2/catalog", brokerPort), nil)
				Expect(err).NotTo(HaveOccurred())

				req.SetBasicAuth(username, password)

				rsp, err = brokerHttpClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusOK))
				if rsp.StatusCode != http.StatusOK {
					Fail(fmt.Sprintf("Not ok:%d", rsp.StatusCode))
				}

				bodyBytes, err := io.ReadAll(rsp.Body)

				FailOnError("Read failed:", err)
				if len(bodyBytes) == 0 {
					Fail("body empty")
				}
				Expect(err).NotTo(HaveOccurred())
				Expect(string(bodyBytes)).To(MatchJSON(catalogBytes))
			})
		})
	})

	Describe("Pubic API", func() {
		AfterEach(func() {
			runner.Interrupt()
			Eventually(runner.Session, 5).Should(Exit(0))
		})
		Context("When a request comes to public api info", func() {
			BeforeEach(func() {
				runner.Start()
			})
			It("succeeds with a 200", func() {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://127.0.0.1:%d/v1/info", publicApiPort), nil)
				Expect(err).NotTo(HaveOccurred())

				rsp, err = apiHttpClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusOK))

				bodyBytes, err := io.ReadAll(rsp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(bodyBytes)).To(MatchJSON(infoBytes))
			})
		})
	})

	Describe("when Health server is ready to serve RESTful API", func() {
		BeforeEach(func() {
			basicAuthConfig := cfg
			basicAuthConfig.Health.HealthCheckUsername = ""
			basicAuthConfig.Health.HealthCheckPassword = ""
			basicAuthConfig.Health.ReadinessCheckEnabled = true
			basicAuthConfig.Health.UnprotectedEndpoints = []string{"/", routes.LivenessPath,
				routes.ReadinessPath, routes.PprofPath, routes.PrometheusPath}
			runner.configPath = writeConfig(&basicAuthConfig).Name()
			runner.Start()
		})
		AfterEach(func() {
			runner.Interrupt()
			Eventually(runner.Session, 5).Should(Exit(0))
		})
		Context("when a request to query health comes", func() {
			It("returns with a 200", func() {
				url := fmt.Sprintf("http://127.0.0.1:%d%s", healthport, routes.PrometheusPath)
				rsp, err := healthHttpClient.Get(url)
				Expect(err).NotTo(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusOK))
				raw, _ := io.ReadAll(rsp.Body)
				healthData := string(raw)
				Expect(healthData).To(ContainSubstring("autoscaler_golangapiserver_concurrent_http_request"))
				Expect(healthData).To(ContainSubstring("autoscaler_golangapiserver_policyDB"))
				Expect(healthData).To(ContainSubstring("autoscaler_golangapiserver_bindingDB"))
				Expect(healthData).To(ContainSubstring("go_goroutines"))
				Expect(healthData).To(ContainSubstring("go_memstats_alloc_bytes"))
				rsp.Body.Close()
			})
		})
	})

	Describe("when Health server is ready to serve RESTful API with basic Auth", func() {
		BeforeEach(func() {
			basicAuthConfig := cfg
			basicAuthConfig.Health.HealthCheckUsername = "correct_username"
			basicAuthConfig.Health.HealthCheckPassword = "correct_password"
			// basicAuthConfig.Health.ReadinessCheckEnabled = true
			// basicAuthConfig.Health.UnprotectedEndpoints = []string{"/", healthendpoint.LIVELINESS_PATH,
			// 	healthendpoint.READINESS_PATH, healthendpoint.PPROF_PATH, healthendpoint.PROMETHEUS_PATH}
			runner.configPath = writeConfig(&basicAuthConfig).Name()
			runner.Start()
		})
		AfterEach(func() {
			runner.Interrupt()
			Eventually(runner.Session, 5).Should(Exit(0))
		})
		Context("when username and password are incorrect for basic authentication during health check", func() {
			It("should return 401", func() {
				url := fmt.Sprintf("http://127.0.0.1:%d%s", healthport, routes.LivenessPath)
				req, err := http.NewRequest(http.MethodGet, url, nil)
				Expect(err).NotTo(HaveOccurred())

				req.SetBasicAuth("wrong_username", "wrong_password")

				rsp, err := healthHttpClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when username and password are correct for basic authentication during health check", func() {
			It("should return 200", func() {
				url := fmt.Sprintf("http://127.0.0.1:%d%s", healthport, routes.LivenessPath)
				req, err := http.NewRequest(http.MethodGet, url, nil)
				Expect(err).NotTo(HaveOccurred())

				req.SetBasicAuth(cfg.Health.HealthCheckUsername, cfg.Health.HealthCheckPassword)

				rsp, err := healthHttpClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("can start with default plugin", func() {
		BeforeEach(func() {
			pluginPathConfig := cfg
			pluginPathConfig.CredHelperImpl = "default"
			runner.configPath = writeConfig(&pluginPathConfig).Name()
			runner.Start()
		})
		AfterEach(func() {
			runner.Interrupt()
			Eventually(runner.Session, 5).Should(Exit(0))
		})
		Context("when a request to query health comes", func() {
			It("returns with a 200", func() {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://127.0.0.1:%d/v1/info", publicApiPort), nil)
				Expect(err).NotTo(HaveOccurred())

				rsp, err = apiHttpClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusOK))

				bodyBytes, err := io.ReadAll(rsp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(bodyBytes)).To(MatchJSON(infoBytes))
			})
		})
	})

})
