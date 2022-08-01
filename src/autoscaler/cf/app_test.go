package cf_test

import (
	"errors"
	"regexp"
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/cf"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"
	. "code.cloudfoundry.org/app-autoscaler/src/autoscaler/testhelpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"encoding/json"
	"net/http"
)

var _ = Describe("Cf client App", func() {

	var (
		conf            *cf.Config
		cfc             cf.CFClient
		fakeCC          *MockServer
		fakeLoginServer *Server
		err             error
		logger          lager.Logger
	)

	var setCfcClient = func(maxRetries int) {
		conf = &cf.Config{}
		conf.API = fakeCC.URL()
		conf.MaxRetries = maxRetries
		conf.MaxRetryWaitMs = 1
		cfc = cf.NewCFClient(conf, logger, clock.NewClock())
		err = cfc.Login()
		Expect(err).NotTo(HaveOccurred())
	}

	BeforeEach(func() {
		fakeCC = NewMockServer()
		fakeLoginServer = NewServer()
		fakeCC.Add().Info(fakeLoginServer.URL())
		fakeLoginServer.RouteToHandler("POST", cf.PathCFAuth, RespondWithJSONEncoded(http.StatusOK, cf.Tokens{
			AccessToken: "test-access-token",
			ExpiresIn:   12000,
		}))
		logger = lager.NewLogger("cf")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		setCfcClient(0)
	})

	AfterEach(func() {
		if fakeCC != nil {
			fakeCC.Close()
		}
		if fakeLoginServer != nil {
			fakeLoginServer.Close()
		}
	})

	Describe("GetApp", func() {

		When("the mocks are used", func() {
			var mocks = NewMockServer()
			BeforeEach(func() {
				conf.API = mocks.URL()
				mocks.Add().GetApp("STARTED").Info(fakeLoginServer.URL())

				DeferCleanup(mocks.Close)
			})
			It("will return success", func() {
				Expect(err).NotTo(HaveOccurred())
				created, err := time.Parse(time.RFC3339, "2022-07-21T13:42:30Z")
				Expect(err).NotTo(HaveOccurred())
				updated, err := time.Parse(time.RFC3339, "2022-07-21T14:30:17Z")
				Expect(err).NotTo(HaveOccurred())
				app, err := cfc.GetApp("test-app-id")
				Expect(err).NotTo(HaveOccurred())
				Expect(app).To(Equal(&cf.App{
					Guid:      "testing-guid-get-app",
					Name:      "mock-get-app",
					State:     "STARTED",
					CreatedAt: created,
					UpdatedAt: updated,
					Relationships: cf.Relationships{
						Space: &cf.Space{
							Data: cf.SpaceData{
								Guid: "test_space_guid",
							},
						},
					},
				}))
			})
		})

		When("get app succeeds", func() {
			BeforeEach(func() {
				fakeCC.AppendHandlers(
					CombineHandlers(
						VerifyRequest("GET", "/v3/apps/test-app-id"),
						RespondWith(http.StatusOK, LoadFile("testdata/app.json"), http.Header{"Content-Type": []string{"application/json"}}),
					),
				)
			})

			It("returns correct state", func() {
				app, err := cfc.GetApp("test-app-id")
				Expect(err).NotTo(HaveOccurred())
				created, err := time.Parse(time.RFC3339, "2022-07-21T13:42:30Z")
				Expect(err).NotTo(HaveOccurred())
				updated, err := time.Parse(time.RFC3339, "2022-07-21T14:30:17Z")
				Expect(err).NotTo(HaveOccurred())
				Expect(app).To(Equal(&cf.App{
					Guid:      "663e9a25-30ba-4fb4-91fa-9b784f4a8542",
					Name:      "autoscaler-1--0cde0e473e3e47f4",
					State:     "STOPPED",
					CreatedAt: created,
					UpdatedAt: updated,
					Relationships: cf.Relationships{
						Space: &cf.Space{
							Data: cf.SpaceData{
								Guid: "3dfc4a10-6e70-44f8-989d-b3842f339e3b",
							},
						},
					},
				}))
			})
		})

		When("get app usage return 404 status code", func() {
			BeforeEach(func() {
				fakeCC.AppendHandlers(
					CombineHandlers(
						VerifyRequest("GET", "/v3/apps/404"),
						RespondWithJSONEncoded(http.StatusNotFound, models.CfResourceNotFound),
					),
				)
			})

			It("should error", func() {
				app, err := cfc.GetApp("404")
				Expect(app).To(BeNil())
				var cfError *models.CfError
				Expect(errors.As(err, &cfError) && cfError.IsNotFound()).To(BeTrue())
				Expect(models.IsNotFound(err)).To(BeTrue())
			})
		})

		When("get app returns 500 status code", func() {
			BeforeEach(func() {
				setCfcClient(3)
			})
			When("it never recovers", func() {

				BeforeEach(func() {
					fakeCC.RouteToHandler("GET", regexp.MustCompile(`^/v3/apps/[^/]+$`),
						RespondWithJSONEncoded(http.StatusInternalServerError, models.CfInternalServerError),
					)
				})

				It("should error", func() {
					app, err := cfc.GetApp("500")
					Expect(app).To(BeNil())
					Expect(fakeCC.Count().Requests(`^/v3/apps/[^/]+$`)).To(Equal(4))
					Expect(err).To(MatchError(MatchRegexp("failed getting app '500':.*'UnknownError'")))
				})
			})
			When("it recovers after 3 retries", func() {
				BeforeEach(func() {
					fakeCC.RouteToHandler("GET", regexp.MustCompile(`^/v3/apps/[^/]+$`),
						RespondWithMultiple(
							RespondWithJSONEncoded(http.StatusInternalServerError, models.CfInternalServerError),
							RespondWithJSONEncoded(http.StatusInternalServerError, models.CfInternalServerError),
							RespondWithJSONEncoded(http.StatusInternalServerError, models.CfInternalServerError),
							RespondWith(http.StatusOK, LoadFile("testdata/app.json"), http.Header{"Content-Type": []string{"application/json"}}),
						))
				})

				It("should return success", func() {
					app, err := cfc.GetApp("500")
					Expect(err).NotTo(HaveOccurred())
					Expect(app).ToNot(BeNil())
					Expect(fakeCC.Count().Requests(`^/v3/apps/[^/]+$`)).To(Equal(4))
				})
			})
		})

		When("get app returns a non-200 and non-404 status code with non-JSON response", func() {
			BeforeEach(func() {
				fakeCC.RouteToHandler("GET", "/v3/apps/invalid_json", RespondWithJSONEncoded(http.StatusInternalServerError, ""))
			})

			It("should error", func() {
				app, err := cfc.GetApp("invalid_json")
				Expect(app).To(BeNil())
				Expect(err.Error()).To(MatchRegexp("failed getting app '.*':.*failed to unmarshal"))
			})
		})

		When("cloud controller is not reachable", func() {
			BeforeEach(func() {
				fakeCC.Close()
				fakeCC = nil
			})

			It("should error", func() {
				app, err := cfc.GetApp("something")
				Expect(app).To(BeNil())
				IsUrlNetOpError(err)
			})
		})

		When("cloud controller returns incorrect message body", func() {
			BeforeEach(func() {
				fakeCC.AppendHandlers(
					CombineHandlers(
						VerifyRequest("GET", "/v3/apps/incorrect_object"),
						RespondWithJSONEncoded(http.StatusOK, `{"entity":{"instances:"abc"}}`),
					),
				)
			})

			It("should error", func() {
				app, err := cfc.GetApp("incorrect_object")
				Expect(app).To(BeNil())
				Expect(err).To(MatchError(MatchRegexp("failed unmarshalling app information for 'incorrect_object': .* cannot unmarshal string")))
				var errType *json.UnmarshalTypeError
				Expect(errors.As(err, &errType)).Should(BeTrue(), "Error was: %#v", interface{}(err))
			})

		})
	})

	Describe("GetAppProcesses", func() {

		When("the mocks are used", func() {
			var mocks = NewMockServer()
			BeforeEach(func() {
				conf.API = mocks.URL()
				mocks.Add().GetAppProcesses(27).Info(fakeLoginServer.URL())
				DeferCleanup(mocks.Close)
			})
			It("will return success", func() {
				app, err := cfc.GetAppProcesses("test-app-id")
				Expect(err).NotTo(HaveOccurred())
				Expect(app).To(Equal(cf.Processes{{Instances: 27}}))
			})
		})

		When("get process with one page succeeds", func() {
			BeforeEach(func() {
				fakeCC.AppendHandlers(
					CombineHandlers(
						VerifyRequest("GET", "/v3/apps/test-app-id/processes", "per_page=100"),
						RespondWith(http.StatusOK, LoadFile("testdata/app_processes.json"), http.Header{"Content-Type": []string{"application/json"}}),
					),
				)
			})

			It("returns correct state", func() {
				processes, err := cfc.GetAppProcesses("test-app-id")
				Expect(err).NotTo(HaveOccurred())
				created, err := time.Parse(time.RFC3339, "2016-03-23T18:48:22Z")
				Expect(err).NotTo(HaveOccurred())
				updated, err := time.Parse(time.RFC3339, "2016-03-23T18:48:42Z")
				Expect(err).NotTo(HaveOccurred())
				Expect(processes).To(Equal(cf.Processes{
					{
						Guid:       "6a901b7c-9417-4dc1-8189-d3234aa0ab82",
						Type:       "web",
						Instances:  5,
						MemoryInMb: 256,
						DiskInMb:   1024,
						CreatedAt:  created,
						UpdatedAt:  updated,
					},
					{
						Guid:       "3fccacd9-4b02-4b96-8d02-8e865865e9eb",
						Type:       "worker",
						Instances:  1,
						MemoryInMb: 256,
						DiskInMb:   1024,
						CreatedAt:  created,
						UpdatedAt:  updated,
					},
				}))
				Expect(processes.GetInstances()).To(Equal(6))
			})
		})

		Context("get process with mutltiple pages", func() {
			type processesResponse struct {
				Pagination cf.Pagination `json:"pagination"`
				Resources  cf.Processes  `json:"resources"`
			}
			When("there are 3 pages", func() {

				BeforeEach(func() {
					fakeCC.AppendHandlers(
						CombineHandlers(
							VerifyRequest("GET", "/v3/apps/test-app-id/processes"),
							RespondWithJSONEncoded(http.StatusOK,
								processesResponse{
									Resources:  cf.Processes{{Instances: 1}, {Instances: 1}},
									Pagination: cf.Pagination{Next: cf.Href{Url: fakeCC.URL() + "/v3/apps/test-app-id/processes/1"}},
								}),
						),
						CombineHandlers(
							VerifyRequest("GET", "/v3/apps/test-app-id/processes/1"),
							RespondWithJSONEncoded(http.StatusOK,
								processesResponse{
									Resources:  cf.Processes{{Instances: 1}, {Instances: 1}},
									Pagination: cf.Pagination{Next: cf.Href{Url: fakeCC.URL() + "/v3/apps/test-app-id/processes/2"}},
								}),
						),
						CombineHandlers(
							VerifyRequest("GET", "/v3/apps/test-app-id/processes/2"),
							RespondWithJSONEncoded(http.StatusOK,
								processesResponse{
									Resources: cf.Processes{{Instances: 1}, {Instances: 1}}},
							),
						),
					)
				})

				It("counts all processes", func() {
					processes, err := cfc.GetAppProcesses("test-app-id")
					Expect(err).ToNot(HaveOccurred())
					Expect(processes.GetInstances()).To(Equal(6))
				})
			})
			When("the second page fails", func() {
				type processesResponse struct {
					Pagination cf.Pagination `json:"pagination"`
					Resources  cf.Processes  `json:"resources"`
				}
				BeforeEach(func() {
					fakeCC.AppendHandlers(
						CombineHandlers(
							VerifyRequest("GET", "/v3/apps/test-app-id/processes"),
							RespondWithJSONEncoded(http.StatusOK,
								processesResponse{
									Resources:  cf.Processes{{Instances: 1}, {Instances: 1}},
									Pagination: cf.Pagination{Next: cf.Href{Url: fakeCC.URL() + "/v3/apps/test-app-id/processes/1"}},
								}),
						),
						CombineHandlers(
							VerifyRequest("GET", "/v3/apps/test-app-id/processes/1"),
							RespondWithJSONEncoded(http.StatusOK,
								processesResponse{
									Resources:  cf.Processes{{Instances: 1}, {Instances: 1}},
									Pagination: cf.Pagination{Next: cf.Href{Url: fakeCC.URL() + "/v3/apps/test-app-id/processes/2"}},
								}),
						),
					)
					fakeCC.RouteToHandler("GET", "/v3/apps/test-app-id/processes/2", RespondWithJSONEncoded(http.StatusInternalServerError, models.CfInternalServerError))
				})

				It("returns correct state", func() {
					_, err := cfc.GetAppProcesses("test-app-id")
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(MatchRegexp("failed getting processes page 3: failed getting processes for app 'test-app-id':.*'UnknownError'.*")))
				})
			})
		})

		When("get processes return 404 status code", func() {
			BeforeEach(func() {
				fakeCC.AppendHandlers(
					CombineHandlers(
						VerifyRequest("GET", "/v3/apps/404/processes"),
						RespondWithJSONEncoded(http.StatusNotFound, models.CfResourceNotFound),
					),
				)
			})

			It("should error", func() {
				process, err := cfc.GetAppProcesses("404")
				Expect(process).To(BeNil())
				var cfError *models.CfError
				Expect(errors.As(err, &cfError) && cfError.IsNotFound()).To(BeTrue())
				Expect(models.IsNotFound(err)).To(BeTrue())
			})
		})

		When("get app returns 500 status code", func() {
			BeforeEach(func() {
				setCfcClient(3)
			})
			When("it never recovers", func() {

				BeforeEach(func() {
					fakeCC.RouteToHandler("GET", "/v3/apps/500/processes",
						RespondWithJSONEncoded(http.StatusInternalServerError, models.CfInternalServerError))
				})

				It("should error", func() {
					process, err := cfc.GetAppProcesses("500")
					Expect(process).To(BeNil())
					Expect(fakeCC.Count().Requests(`^/v3/apps/500/processes\?.*`)).To(Equal(4))
					Expect(err).To(MatchError(MatchRegexp("failed getting processes for app '500':.*'UnknownError'")))
				})
			})
			When("it recovers after 3 retries", func() {
				BeforeEach(func() {
					fakeCC.RouteToHandler("GET", "/v3/apps/500/processes",
						RespondWithMultiple(
							RespondWithJSONEncoded(http.StatusInternalServerError, models.CfInternalServerError),
							RespondWithJSONEncoded(http.StatusInternalServerError, models.CfInternalServerError),
							RespondWithJSONEncoded(http.StatusInternalServerError, models.CfInternalServerError),
							RespondWith(http.StatusOK, LoadFile("testdata/app_processes.json"), http.Header{"Content-Type": []string{"application/json"}}),
						))
				})

				It("should return success", func() {
					process, err := cfc.GetAppProcesses("500")
					Expect(err).To(BeNil())
					Expect(fakeCC.Count().Requests(`^/v3/apps/500/processes\?`)).To(Equal(4))
					Expect(process).ToNot(BeNil())
				})
			})
		})

		When("get processes returns a non-200 and non-404 status code with non-JSON response", func() {
			BeforeEach(func() {
				fakeCC.AppendHandlers(
					CombineHandlers(
						VerifyRequest("GET", "/v3/apps/invalid_json/processes"),
						RespondWithJSONEncoded(http.StatusInternalServerError, ""),
					),
				)
			})

			It("should error", func() {
				process, err := cfc.GetAppProcesses("invalid_json")
				Expect(process).To(BeNil())
				Expect(err.Error()).To(MatchRegexp("failed getting processes for app '.*':.*failed to unmarshal"))
			})
		})

		When(" get processes call and cloud controller is not reachable", func() {
			BeforeEach(func() {
				fakeCC.Close()
				fakeCC = nil
			})

			It("should error", func() {
				app, err := cfc.GetAppProcesses("something")
				Expect(app).To(BeNil())
				IsUrlNetOpError(err)
			})
		})

		When("get processes returns incorrect message body", func() {
			BeforeEach(func() {
				fakeCC.AppendHandlers(
					CombineHandlers(
						VerifyRequest("GET", "/v3/apps/incorrect_object/processes"),
						RespondWithJSONEncoded(http.StatusOK, `{"entity":{"instances:"abc"}}`),
					),
				)
			})

			It("should error", func() {
				process, err := cfc.GetAppProcesses("incorrect_object")
				Expect(process).To(BeNil())
				Expect(err).To(MatchError(MatchRegexp("failed unmarshalling processes information for 'incorrect_object': .* cannot unmarshal string")))
				var errType *json.UnmarshalTypeError
				Expect(errors.As(err, &errType)).Should(BeTrue(), "Error was: %#v", interface{}(err))
			})

		})
	})

	Describe("GetStateAndInstances", func() {

		When("the mocks are used", func() {
			var mocks = NewMockServer()
			BeforeEach(func() {
				conf.API = mocks.URL()
				mocks.Add().GetAppProcesses(27).Info(fakeLoginServer.URL())
				mocks.Add().GetApp("STARTED")
				DeferCleanup(mocks.Close)
			})
			It("will return success", func() {
				app, err := cfc.GetStateAndInstances("test-app-id")
				Expect(err).NotTo(HaveOccurred())
				state := "STARTED"
				Expect(app).To(Equal(&models.AppEntity{State: &state, Instances: 27}))
			})
		})

		When("get app & process return ok", func() {
			BeforeEach(func() {
				fakeCC.RouteToHandler("GET", "/v3/apps/test-app-id/processes", CombineHandlers(
					RespondWith(http.StatusOK, LoadFile("testdata/app_processes.json"), http.Header{"Content-Type": []string{"application/json"}}),
				))
				fakeCC.RouteToHandler("GET", "/v3/apps/test-app-id", CombineHandlers(
					RespondWith(http.StatusOK, LoadFile("testdata/app.json"), http.Header{"Content-Type": []string{"application/json"}}),
				))
			})

			It("returns correct state", func() {
				processes, err := cfc.GetStateAndInstances("test-app-id")
				Expect(err).NotTo(HaveOccurred())
				s := "STOPPED"
				Expect(processes).To(Equal(&models.AppEntity{Instances: 6, State: &s}))
			})
		})

		When("get app returns 500 & get process return ok", func() {
			BeforeEach(func() {
				fakeCC.RouteToHandler("GET", "/v3/apps/test-app-id/processes", CombineHandlers(
					RespondWithJSONEncoded(http.StatusInternalServerError, models.CfInternalServerError),
				))
				fakeCC.RouteToHandler("GET", "/v3/apps/test-app-id", CombineHandlers(
					RespondWith(http.StatusOK, LoadFile("testdata/app.json"), http.Header{"Content-Type": []string{"application/json"}}),
				))
			})

			It("should error", func() {
				entity, err := cfc.GetStateAndInstances("test-app-id")
				Expect(entity).To(BeNil())
				Expect(err).To(MatchError(MatchRegexp("get state&instances GetAppProcesses failed: failed getting processes.*:.*'UnknownError'")))
			})
		})

		When("get processes return OK get app returns 500", func() {
			BeforeEach(func() {
				fakeCC.RouteToHandler("GET", "/v3/apps/test-app-id/processes", CombineHandlers(
					RespondWith(http.StatusOK, LoadFile("testdata/app_processes.json"), http.Header{"Content-Type": []string{"application/json"}}),
				))
				fakeCC.RouteToHandler("GET", "/v3/apps/test-app-id", CombineHandlers(
					RespondWithJSONEncoded(http.StatusInternalServerError, models.CfInternalServerError),
				))
			})

			It("should error", func() {
				entity, err := cfc.GetStateAndInstances("test-app-id")
				Expect(entity).To(BeNil())
				Expect(err).To(MatchError(MatchRegexp("get state&instances getApp failed:.*failed getting app.*:.*'UnknownError'")))
			})
		})

		When("get processes return 500 & get app returns 500", func() {
			BeforeEach(func() {
				fakeCC.RouteToHandler("GET", "/v3/apps/test-app-id/processes", CombineHandlers(
					RespondWithJSONEncoded(http.StatusInternalServerError, models.CfInternalServerError),
				))
				fakeCC.RouteToHandler("GET", "/v3/apps/test-app-id", CombineHandlers(
					RespondWithJSONEncoded(http.StatusInternalServerError, models.CfInternalServerError),
				))
			})

			It("should error", func() {
				entity, err := cfc.GetStateAndInstances("test-app-id")
				Expect(entity).To(BeNil())
				Expect(err).To(MatchError(MatchRegexp("get state&instances getApp failed: failed getting app.*:.*'UnknownError'")))
			})
		})
	})

	Describe("SetAppInstances", func() {
		JustBeforeEach(func() {
			err = cfc.SetAppInstances("test-app-id", 6)
		})
		Context("when set app instances succeeds", func() {
			BeforeEach(func() {
				fakeCC.AppendHandlers(
					CombineHandlers(
						VerifyRequest("PUT", cf.PathApp+"/test-app-id"),
						VerifyJSONRepresenting(models.AppEntity{Instances: 6}),
						RespondWith(http.StatusCreated, ""),
					),
				)
			})

			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when updating app instances returns non-200 status code", func() {
			BeforeEach(func() {
				responseMap := make(map[string]interface{})
				responseMap["description"] = "You have exceeded the instance memory limit for your space's quota"
				responseMap["error_code"] = "SpaceQuotaInstanceMemoryLimitExceeded"
				fakeCC.AppendHandlers(
					CombineHandlers(
						RespondWithJSONEncoded(http.StatusBadRequest, responseMap),
					),
				)
			})

			It("should error", func() {
				Expect(err).To(MatchError(MatchRegexp("failed setting application instances: *")))
			})

		})

		Context("when cloud controller is not reachable", func() {
			BeforeEach(func() {
				ccURL := fakeCC.URL()
				fakeCC.Close()
				fakeCC = nil

				Eventually(func() error {
					// #nosec G107
					resp, err := http.Get(ccURL)

					if err != nil {
						return err
					}
					_ = resp.Body.Close()

					return nil
				}).Should(HaveOccurred())
			})

			It("should error", func() {
				IsUrlNetOpError(err)
			})

		})

	})

})
