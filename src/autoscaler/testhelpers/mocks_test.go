package testhelpers_test

import (
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/cf"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"
	. "code.cloudfoundry.org/app-autoscaler/src/autoscaler/testhelpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"net/http"
)

var _ = Describe("Cf cloud controller", func() {

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

})
