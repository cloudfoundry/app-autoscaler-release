package main_test

import (
	"io"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/cf"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/onsi/gomega/gbytes"

	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

var _ = Describe("Main", func() {

	var (
		runner    *ScalingEngineRunner
		serverURL string
	)

	BeforeEach(func() {
		runner = NewScalingEngineRunner()
		serverURL = fmt.Sprintf("https://127.0.0.1:%d", conf.Server.Port)
	})

	JustBeforeEach(func() {
		runner.Start()
	})

	AfterEach(func() {
		runner.KillWithFire()
	})

	Describe("with a correct config", func() {

		Context("when starting 1 scaling engine instance", func() {
			It("scaling engine should start", func() {
				Eventually(runner.Session.Buffer, 2*time.Second).Should(gbytes.Say(runner.startCheck))
				Consistently(runner.Session).ShouldNot(Exit())
			})

			It("http server starts directly", func() {
				Eventually(runner.Session.Buffer, 2*time.Second).Should(gbytes.Say("scalingengine.http-server.new-http-server"))
			})
		})

		Context("when starting multiple scaling engine instances", func() {
			var (
				secondRunner *ScalingEngineRunner
			)

			JustBeforeEach(func() {
				secondRunner = NewScalingEngineRunner()
				secondConf := conf

				secondConf.Server.Port += 500
				secondConf.Health.Port += 500
				secondRunner.configPath = writeConfig(&secondConf).Name()
				secondRunner.Start()
			})

			AfterEach(func() {
				secondRunner.KillWithFire()
			})

			It("2 http server instances start", func() {
				Eventually(runner.Session.Buffer, 2*time.Second).Should(gbytes.Say("scalingengine.http-server.new-http-server"))
				Eventually(secondRunner.Session.Buffer, 2*time.Second).Should(gbytes.Say("scalingengine.http-server.new-http-server"))
				Eventually(runner.Session.Buffer, 2*time.Second).Should(gbytes.Say("scalingengine.started"))
				Eventually(secondRunner.Session.Buffer, 2*time.Second).Should(gbytes.Say("scalingengine.started"))

				Consistently(runner.Session).ShouldNot(Exit())
				Consistently(secondRunner.Session).ShouldNot(Exit())
			})

		})

	})

	Describe("With incorrect config", func() {

		Context("with a missing config file", func() {
			BeforeEach(func() {
				runner.startCheck = ""
				runner.configPath = "bogus"
			})

			It("fails with an error", func() {
				Eventually(runner.Session).Should(Exit(1))
				Expect(runner.Session.Buffer()).To(gbytes.Say("failed to open config file"))
			})
		})

		Context("with an invalid config file", func() {
			BeforeEach(func() {
				runner.startCheck = ""
				badfile, err := os.CreateTemp("", "bad-engine-config")
				Expect(err).NotTo(HaveOccurred())
				runner.configPath = badfile.Name()
				// #nosec G306
				err = os.WriteFile(runner.configPath, []byte("bogus"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				_ = os.Remove(runner.configPath)
			})

			It("fails with an error", func() {
				Eventually(runner.Session).Should(Exit(1))
				Expect(runner.Session.Buffer()).To(gbytes.Say("failed to read config file"))
			})
		})

		Context("with missing configuration", func() {
			BeforeEach(func() {
				runner.startCheck = ""
				missingParamConf := conf
				missingParamConf.CF = cf.Config{
					API: ccUAA.URL(),
				}

				missingParamConf.Server.Port = 7000 + GinkgoParallelProcess()
				missingParamConf.Logging.Level = "debug"

				cfg := writeConfig(&missingParamConf)
				runner.configPath = cfg.Name()
			})

			AfterEach(func() {
				os.Remove(runner.configPath)
			})

			It("should fail validation", func() {
				Eventually(runner.Session).Should(Exit(1))
				Expect(runner.Session.Buffer()).To(gbytes.Say("failed to validate configuration"))
			})
		})
	})

	Describe("when http server is ready to serve RESTful API", func() {

		JustBeforeEach(func() {
			Eventually(runner.Session.Buffer, 2).Should(gbytes.Say("scalingengine.started"))
		})

		Context("when a request to trigger scaling comes", func() {
			It("returns with a 200", func() {
				body, err := json.Marshal(models.Trigger{Adjustment: "+1"})
				Expect(err).NotTo(HaveOccurred())

				rsp, err := httpClient.Post(fmt.Sprintf("%s/v1/apps/%s/scale", serverURL, appId),
					"application/json", bytes.NewReader(body))
				Expect(err).NotTo(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusOK))
				rsp.Body.Close()
			})
		})

		Context("when a request to retrieve scaling history comes", func() {
			It("returns with a 200", func() {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v1/apps/%s/scaling_histories", serverURL, appId), nil)
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Authorization", "Bearer none")
				rsp, err := httpClient.Do(req)
				Expect(err).NotTo(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusOK))
				rsp.Body.Close()
			})
		})

		It("handles the start and end of a schedule", func() {
			By("start of a schedule")
			url := fmt.Sprintf("%s/v1/apps/%s/active_schedules/111111", serverURL, appId)
			bodyReader := bytes.NewReader([]byte(`{"instance_min_count":1, "instance_max_count":5, "initial_min_instance_count":3}`))

			req, err := http.NewRequest(http.MethodPut, url, bodyReader)
			Expect(err).NotTo(HaveOccurred())

			rsp, err := httpClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(rsp.StatusCode).To(Equal(http.StatusOK))
			rsp.Body.Close()

			By("end of a schedule")
			req, err = http.NewRequest(http.MethodDelete, url, nil)
			Expect(err).NotTo(HaveOccurred())

			rsp, err = httpClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(rsp.StatusCode).To(Equal(http.StatusOK))
			rsp.Body.Close()
		})
	})

	Describe("when Health server is ready to serve RESTful API", func() {
		BeforeEach(func() {
			basicAuthConfig := conf
			basicAuthConfig.Health.HealthCheckUsername = ""
			basicAuthConfig.Health.HealthCheckPassword = ""
			runner.configPath = writeConfig(&basicAuthConfig).Name()
		})

		JustBeforeEach(func() {
			Eventually(runner.Session.Buffer, 2).Should(gbytes.Say("scalingengine.started"))
		})

		Context("when a request to query health comes", func() {
			It("returns with a 200", func() {
				rsp, err := httpClient.Get(serverURL)
				Expect(err).NotTo(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusOK))
				raw, _ := io.ReadAll(rsp.Body)
				healthData := string(raw)
				Expect(healthData).To(ContainSubstring("autoscaler_scalingengine_concurrent_http_request"))
				Expect(healthData).To(ContainSubstring("autoscaler_scalingengine_schedulerDB"))
				Expect(healthData).To(ContainSubstring("autoscaler_scalingengine_policyDB"))
				Expect(healthData).To(ContainSubstring("autoscaler_scalingengine_scalingengineDB"))
				Expect(healthData).To(ContainSubstring("go_goroutines"))
				Expect(healthData).To(ContainSubstring("go_memstats_alloc_bytes"))
				rsp.Body.Close()

			})
		})
	})

	Describe("when Health server is ready to serve RESTful API with basic Auth", func() {
		JustBeforeEach(func() {
			Eventually(runner.Session.Buffer, 2).Should(gbytes.Say("scalingengine.started"))
		})

		Context("when username and password are incorrect for basic authentication during health check", func() {
			It("should return 401", func() {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/health", serverURL), nil)
				Expect(err).NotTo(HaveOccurred())

				req.SetBasicAuth("wrongusername", "wrongpassword")

				rsp, err := httpClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when username and password are correct for basic authentication during health check", func() {
			It("should return 200", func() {

				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/health", serverURL), nil)
				Expect(err).NotTo(HaveOccurred())

				req.SetBasicAuth(conf.Health.HealthCheckUsername, conf.Health.HealthCheckPassword)

				rsp, err := httpClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("when Health server is ready to serve RESTful API with basic Auth", func() {
		JustBeforeEach(func() {
			Eventually(runner.Session.Buffer, 2).Should(gbytes.Say("scalingengine.started"))
		})

		Context("when username and password are incorrect for basic authentication during health check", func() {
			It("should return 401", func() {

				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/health", serverURL), nil)
				Expect(err).NotTo(HaveOccurred())

				req.SetBasicAuth("wrongusername", "wrongpassword")

				rsp, err := httpClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when username and password are correct for basic authentication during health check", func() {
			It("should return 200", func() {

				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/health", serverURL), nil)
				Expect(err).NotTo(HaveOccurred())

				req.SetBasicAuth(conf.Health.HealthCheckUsername, conf.Health.HealthCheckPassword)

				rsp, err := httpClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})
})
