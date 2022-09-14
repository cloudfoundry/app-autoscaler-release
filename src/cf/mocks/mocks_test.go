package mocks_test

import (
	"errors"
	. "github.com/cloudfoundry/app-autoscaler-release/cf"
	. "github.com/cloudfoundry/app-autoscaler-release/cf/mocks"
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"github.com/onsi/gomega/ghttp"

	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cf cloud controller", func() {

	var (
		conf            *Config
		cfc             *Client
		fakeCC          *MockServer
		fakeLoginServer *MockServer
		err             error
		logger          lager.Logger
	)

	var setCfcClient = func(maxRetries int, apiUrl string) {
		conf = &Config{}
		conf.API = apiUrl
		conf.MaxRetries = maxRetries
		conf.MaxRetryWaitMs = 1
		cfc = NewCFClient(conf, logger, clock.NewClock())
		err = cfc.Login()
		Expect(err).NotTo(HaveOccurred())
	}

	BeforeEach(func() {
		fakeCC = NewMockServer()
		fakeLoginServer = NewMockServer()
		fakeCC.Add().Info(fakeLoginServer.URL())
		fakeLoginServer.RouteToHandler("POST", PathCFAuth, ghttp.RespondWithJSONEncoded(http.StatusOK, Tokens{
			AccessToken: "test-access-token",
			ExpiresIn:   12000,
		}))
		logger = lager.NewLogger("cf")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))
		setCfcClient(0, fakeCC.URL())
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
				mocks.Add().GetApp("STARTED", http.StatusOK, "test_space_guid").Info(fakeLoginServer.URL())

				DeferCleanup(mocks.Close)
			})
			It("will return success", func() {
				app, err := cfc.GetApp("test-app-id")
				Expect(err).NotTo(HaveOccurred())
				Expect(app).To(Equal(&App{
					Guid:      "testing-guid-get-app",
					Name:      "mock-get-app",
					State:     "STARTED",
					CreatedAt: ParseDate("2022-07-21T13:42:30Z"),
					UpdatedAt: ParseDate("2022-07-21T14:30:17Z"),
					Relationships: Relationships{
						Space: &Space{
							Data: SpaceData{
								Guid: "test_space_guid",
							},
						},
					},
				}))
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
				app, err := cfc.GetAppProcesses("test-app-id", ProcessTypeWeb)
				Expect(err).NotTo(HaveOccurred())
				Expect(app).To(Equal(Processes{{Instances: 27}}))
			})
		})

	})

	Describe("GetAppAndProcesses", func() {
		When("the mocks are used", func() {
			var mocks = NewMockServer()
			BeforeEach(func() {
				conf.API = mocks.URL()
				mocks.Add().GetAppProcesses(27).Info(fakeLoginServer.URL())
				mocks.Add().GetApp("STARTED", http.StatusOK, "test_space_guid")
				DeferCleanup(mocks.Close)
			})
			It("will return success", func() {
				app, err := cfc.GetAppAndProcesses("test-app-id")
				Expect(err).NotTo(HaveOccurred())
				Expect(app).To(Equal(&AppAndProcesses{
					App: &App{
						Guid:      "testing-guid-get-app",
						Name:      "mock-get-app",
						State:     "STARTED",
						CreatedAt: ParseDate("2022-07-21T13:42:30Z"),
						UpdatedAt: ParseDate("2022-07-21T14:30:17Z"),
						Relationships: Relationships{
							Space: &Space{
								Data: SpaceData{
									Guid: "test_space_guid",
								},
							},
						},
					},
					Processes: Processes{{Instances: 27}},
				}))
			})
		})

	})

	Describe("ScaleAppWebProcess", func() {
		JustBeforeEach(func() {
			err = cfc.ScaleAppWebProcess("test-app-id", 6)
		})

		When("the mocks are used", func() {
			var mocks = NewMockServer()
			BeforeEach(func() {
				conf.API = mocks.URL()
				mocks.Add().ScaleAppWebProcess().Info(fakeLoginServer.URL())
				DeferCleanup(mocks.Close)
			})
			It("will return success", func() {
				err := cfc.ScaleAppWebProcess("r_scalingengine:503,testAppId,1:c8ec66ba", 3)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("GetRoles", func() {
		When("the mocks are used", func() {
			var mocks = NewMockServer()
			BeforeEach(func() {
				conf.API = mocks.URL()
				mocks.Add().Info(fakeLoginServer.URL()).Roles(http.StatusOK, Role{Guid: "mock_guid", Type: RoleSpaceDeveloper})

				DeferCleanup(mocks.Close)
			})
			It("will return success", func() {
				roles, err := cfc.GetSpaceDeveloperRoles("some_space", "some_user")
				Expect(err).NotTo(HaveOccurred())
				Expect(roles).To(Equal(Roles{
					{
						Guid: "mock_guid",
						Type: RoleSpaceDeveloper,
					},
				}))
				Expect(roles.HasRole(RoleSpaceDeveloper)).To(BeTrue())
			})
		})
	})

	Describe("GetServiceInstance", func() {
		When("the mocks are used", func() {
			var mocks = NewMockServer()
			BeforeEach(func() {
				conf.API = mocks.URL()
				mocks.Add().Info(fakeLoginServer.URL()).ServiceInstance("A-service-plan-guid")
				DeferCleanup(mocks.Close)
			})
			It("will return success", func() {
				roles, err := cfc.GetServiceInstance("some-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(roles).To(Equal(&ServiceInstance{
					Guid:          "service-instance-mock-guid",
					Type:          "managed",
					Relationships: ServiceInstanceRelationships{ServicePlan: ServicePlanRelation{Data: ServicePlanData{Guid: "A-service-plan-guid"}}}}))
			})
		})
	})

	Describe("ServicePlan", func() {
		When("the mocks are used", func() {
			var mocks = NewMockServer()
			BeforeEach(func() {
				conf.API = mocks.URL()
				mocks.Add().Info(fakeLoginServer.URL()).ServicePlan("a-broker-plan-id")
				DeferCleanup(mocks.Close)
			})
			It("will return success", func() {
				roles, err := cfc.GetServicePlan("a-broker-plan-id")
				Expect(err).NotTo(HaveOccurred())
				Expect(roles).To(Equal(&ServicePlan{BrokerCatalog: BrokerCatalog{Id: "a-broker-plan-id"}}))
			})
		})
	})

	Describe("Info", func() {
		When("the mocks are used", func() {
			var mocks = NewMockServer()
			BeforeEach(func() {
				conf.API = mocks.URL()
				mocks.Add().Info(fakeLoginServer.URL())
				DeferCleanup(mocks.Close)
			})
			It("will return success", func() {
				endpoints, err := cfc.GetEndpoints()
				Expect(err).NotTo(HaveOccurred())
				Expect(endpoints).To(Equal(Endpoints{
					Login: Href{fakeLoginServer.URL()},
					Uaa:   Href{fakeLoginServer.URL()},
				}))
			})
		})
	})

	Describe("OauthToken", func() {
		When("the mocks are used", func() {
			var mocks = NewMockServer()
			BeforeEach(func() {
				mocks.Add().Info(mocks.URL())
				mocks.Add().OauthToken("a-access-token")
				setCfcClient(0, mocks.URL())
				DeferCleanup(mocks.Close)
			})
			It("will return success", func() {

				token, err := cfc.GetTokens()
				Expect(err).NotTo(HaveOccurred())
				Expect(token).To(Equal(Tokens{AccessToken: "a-access-token", ExpiresIn: 12000}))
			})
		})
	})

	Describe("CheckToken", func() {
		When("the mocks are used", func() {
			var mocks = NewMockServer()
			BeforeEach(func() {
				testUserScope := []string{"cloud_controller.admin"}
				mocks.Add().Info(mocks.URL())
				mocks.Add().CheckToken(testUserScope).OauthToken("a-test-access-token")
				setCfcClient(0, mocks.URL())
				DeferCleanup(mocks.Close)
			})
			It("will return success", func() {
				userAdmin, err := cfc.IsUserAdmin("bearer a-test-access-token")
				Expect(err).NotTo(HaveOccurred())
				Expect(userAdmin).To(BeTrue())
			})
		})
	})

	Describe("UserInfo", func() {
		When("the mocks are used", func() {
			var mocks = NewMockServer()
			BeforeEach(func() {
				mocks.Add().
					Info(mocks.URL()).
					GetApp("", 200, "some-space-guid").
					Roles(http.StatusOK, Role{Guid: "mock_guid", Type: RoleSpaceDeveloper}).
					UserInfo(http.StatusOK, "testUser").
					OauthToken("a-test-access-token")
				setCfcClient(0, mocks.URL())
				DeferCleanup(mocks.Close)
			})
			It("will return success", func() {
				userAdmin, err := cfc.IsUserSpaceDeveloper("bearer a-test-access-token", "test-app-id")
				Expect(err).NotTo(HaveOccurred())
				Expect(userAdmin).To(BeTrue())
			})
		})
		When("the mocks return 401", func() {
			var mocks = NewMockServer()
			BeforeEach(func() {
				mocks.Add().
					Info(mocks.URL()).
					GetApp("", 200, "some-space-guid").
					Roles(http.StatusOK, Role{Guid: "mock_guid", Type: RoleSpaceDeveloper}).
					UserInfo(http.StatusUnauthorized, "testUser").
					OauthToken("a-test-access-token")
				setCfcClient(0, mocks.URL())
				DeferCleanup(mocks.Close)
			})
			It("will return unauthorised", func() {
				_, err := cfc.IsUserSpaceDeveloper("bearer a-test-access-token", "test-app-id")
				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, ErrUnauthrorized)).To(BeTrue())
			})
		})
	})
})

func ParseDate(date string) time.Time {
	updated, err := time.Parse(time.RFC3339, date)
	Expect(err).NotTo(HaveOccurred())
	return updated
}
