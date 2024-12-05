package publicapiserver_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ScalingHistoryHandler", func() {
	var scalingEngineHandler http.HandlerFunc

	JustBeforeEach(func() {
		scalingEnginePathMatcher, err := regexp.Compile(`/v1/apps/[A-Za-z0-9\-]+/scaling_histories`)
		Expect(err).NotTo(HaveOccurred())
		scalingEngineServer.RouteToHandler(http.MethodGet, scalingEnginePathMatcher, scalingEngineHandler)
	})

	BeforeEach(func() {
		scalingEngineHandler = ghttp.RespondWithJSONEncodedPtr(&scalingEngineStatus, &scalingEngineResponse)
	})

	Describe("GET /v1/apps/:appId/aggregated_metric_histories/:metricType", func() {
	})
})
