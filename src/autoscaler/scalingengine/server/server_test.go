package server_test

import (
	"strconv"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/fakes"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/routes"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/scalingengine/config"
	. "code.cloudfoundry.org/app-autoscaler/src/autoscaler/scalingengine/server"
	"code.cloudfoundry.org/lager/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon_v2"

	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

var _ = Describe("Server", func() {
	var (
		username string
		password string

		serverUrl       *url.URL
		server          ifrit.Process
		scalingEngineDB *fakes.FakeScalingEngineDB
		sychronizer     *fakes.FakeActiveScheduleSychronizer

		conf *config.Config

		rsp        *http.Response
		req        *http.Request
		body       []byte
		err        error
		method     string
		bodyReader io.Reader
		route      = routes.ScalingEngineRoutes()
	)

	BeforeEach(func() {
		port := 2222 + GinkgoParallelProcess()
		conf = &config.Config{
			Server: helpers.ServerConfig{
				Port: port,
				BasicAuth: models.BasicAuth{
					Username: "scalingengine",
					Password: "scalingengine-password",
				},
			},
		}
		scalingEngineDB = &fakes.FakeScalingEngineDB{}
		scalingEngine := &fakes.FakeScalingEngine{}
		policyDb := &fakes.FakePolicyDB{}
		schedulerDB := &fakes.FakeSchedulerDB{}
		sychronizer = &fakes.FakeActiveScheduleSychronizer{}

		httpServer, err := NewServer(lager.NewLogger("test"), conf, policyDb, scalingEngineDB, schedulerDB, scalingEngine, sychronizer)
		Expect(err).NotTo(HaveOccurred())
		server = ginkgomon_v2.Invoke(httpServer)
		serverUrl, err = url.Parse("http://127.0.0.1:" + strconv.Itoa(port))
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		ginkgomon_v2.Interrupt(server)
	})
	JustBeforeEach(func() {
		req, err = http.NewRequest(method, serverUrl.String(), bodyReader)
		req.SetBasicAuth(username, password)
		Expect(err).NotTo(HaveOccurred())
		rsp, err = http.DefaultClient.Do(req)
	})

	When("triggering scaling action", func() {
		BeforeEach(func() {
			body, err = json.Marshal(models.Trigger{Adjustment: "+1"})
			Expect(err).NotTo(HaveOccurred())

			bodyReader = bytes.NewReader(body)
			uPath, err := route.Get(routes.ScaleRouteName).URLPath("appid", "test-app-id")
			Expect(err).NotTo(HaveOccurred())
			serverUrl.Path = uPath.Path
		})

		When("requesting correctly", func() {
			BeforeEach(func() {
				method = http.MethodPost
				username = conf.Server.BasicAuth.Username
				password = conf.Server.BasicAuth.Password
			})

			It("should return 200", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusOK))
				rsp.Body.Close()
			})
		})
	})

	When("getting scaling histories", func() {
		BeforeEach(func() {
			uPath, err := route.Get(routes.GetScalingHistoriesRouteName).URLPath("guid", "8ea70e4e-e0bc-4e15-9d32-cd69daaf012a")
			Expect(err).NotTo(HaveOccurred())
			serverUrl.Path = uPath.Path
		})

		JustBeforeEach(func() {
			serverUrl.User = url.UserPassword(username, password)
			req, err = http.NewRequest(method, serverUrl.String(), nil)
			Expect(err).NotTo(HaveOccurred())

			auth := username + ":" + password
			base64Auth := base64.StdEncoding.EncodeToString([]byte(auth))
			req.Header.Set("Authorization", "Basic "+base64Auth)

			rsp, err = (&http.Client{}).Do(req)
			Expect(err).NotTo(HaveOccurred())
		})

		When("credentials are correct", func() {
			BeforeEach(func() {
				method = http.MethodGet
				username = conf.Server.BasicAuth.Username
				password = conf.Server.BasicAuth.Password
			})

			It("should return 200", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusOK))
				rsp.Body.Close()
			})
		})
	})

	When("requesting active shedule", func() {

		BeforeEach(func() {
			uPath, err := route.Get(routes.SetActiveScheduleRouteName).URLPath("appid", "test-app-id", "scheduleid", "test-schedule-id")
			Expect(err).NotTo(HaveOccurred())
			serverUrl.Path = uPath.Path
		})

		When("setting active schedule", func() {
			BeforeEach(func() {
				bodyReader = bytes.NewReader([]byte(`{"instance_min_count":1, "instance_max_count":5, "initial_min_instance_count":3}`))
			})

			When("credentials are correct", func() {
				BeforeEach(func() {
					method = http.MethodPut
					username = conf.Server.BasicAuth.Username
					password = conf.Server.BasicAuth.Password
				})

				It("should return 200", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(rsp.StatusCode).To(Equal(http.StatusOK))
					rsp.Body.Close()
				})
			})
		})

		When("deleting active schedule", func() {
			BeforeEach(func() {
				uPath, err := route.Get(routes.DeleteActiveScheduleRouteName).URLPath("appid", "test-app-id", "scheduleid", "test-schedule-id")
				Expect(err).NotTo(HaveOccurred())
				serverUrl.Path = uPath.Path
				bodyReader = nil
				method = http.MethodDelete
			})

			When("requesting correctly", func() {

				BeforeEach(func() {
					username = conf.Server.BasicAuth.Username
					password = conf.Server.BasicAuth.Password
				})

				It("should return 200", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(rsp.StatusCode).To(Equal(http.StatusOK))
					rsp.Body.Close()
				})
			})
		})

		When("getting active schedule", func() {
			BeforeEach(func() {
				uPath, err := route.Get(routes.GetActiveSchedulesRouteName).URLPath("appid", "test-app-id")
				Expect(err).NotTo(HaveOccurred())
				serverUrl.Path = uPath.Path
				bodyReader = nil
				method = http.MethodGet
			})

			When("requesting correctly", func() {
				BeforeEach(func() {
					username = conf.Server.BasicAuth.Username
					password = conf.Server.BasicAuth.Password

					activeSchedule := &models.ActiveSchedule{
						ScheduleId:         "a-schedule-id",
						InstanceMin:        1,
						InstanceMax:        5,
						InstanceMinInitial: 3,
					}

					scalingEngineDB.GetActiveScheduleReturns(activeSchedule, nil)
				})

				It("should return 200", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(rsp.StatusCode).To(Equal(http.StatusOK))
					rsp.Body.Close()
				})
			})
		})
	})

	When("requesting sync shedule", func() {
		BeforeEach(func() {
			uPath, err := route.Get(routes.SyncActiveSchedulesRouteName).URLPath()
			Expect(err).NotTo(HaveOccurred())
			serverUrl.Path = uPath.Path
			bodyReader = nil
		})

		When("requesting correctly", func() {
			BeforeEach(func() {
				method = http.MethodPut
				username = conf.Server.BasicAuth.Username
				password = conf.Server.BasicAuth.Password
			})

			It("should return 200", func() {
				Eventually(sychronizer.SyncCallCount).Should(Equal(1))
				Expect(err).ToNot(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusOK))
				rsp.Body.Close()
			})
		})

		When("requesting with incorrect http method", func() {
			BeforeEach(func() {
				method = http.MethodGet
			})

			It("should return 405", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(rsp.StatusCode).To(Equal(http.StatusMethodNotAllowed))
				rsp.Body.Close()
			})
		})

	})
})
