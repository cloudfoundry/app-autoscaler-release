package publicapiserver_test

import (
	"net/http"
	"regexp"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers/auth"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/testhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = XDescribe("ScalingHistoryHandler", func() {
	var scalingEngineHandler http.HandlerFunc

	JustBeforeEach(func() {
		scalingEnginePathMatcher, err := regexp.Compile(`/v1/apps/[A-Za-z0-9\-]+/scaling_histories`)
		Expect(err).NotTo(HaveOccurred())
		scalingEngineServer.RouteToHandler(http.MethodGet, scalingEnginePathMatcher, scalingEngineHandler)
	})

	Describe("GET /v1/apps/:appId/scaling_histories", func() {
		When("conf.CfInstanceCert is set", func() {
			BeforeEach(func() {
				fullCert, err := testhelpers.GenerateClientCert("org-guid", "space-guid")
				Expect(err).NotTo(HaveOccurred())

				cert := auth.NewCert(string(fullCert))
				conf.CfInstanceCert = cert.FullChainPem
				xfccHeaderExpectedValue := cert.GetXFCCHeader()

				scalingEngineHandler = ghttp.CombineHandlers(
					ghttp.VerifyHeader(http.Header{"X-Forwarded-Client-Cert": []string{xfccHeaderExpectedValue}}),
					ghttp.RespondWithJSONEncodedPtr(&eventGeneratorStatus, &eventGeneratorResponse),
				)
			})

			It("should send the XFCC header to the scaling engine", func() {
				req, err := http.NewRequest(http.MethodGet, server.URL()+"/v1/apps/app-id/scaling_histories", nil)
			})
		})
	})
})
