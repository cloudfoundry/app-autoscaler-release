package server

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/app-autoscaler-release/cred_helper"
	"github.com/cloudfoundry/app-autoscaler-release/db"
	"github.com/cloudfoundry/app-autoscaler-release/healthendpoint"
	"github.com/cloudfoundry/app-autoscaler-release/metricsforwarder/config"
	"github.com/cloudfoundry/app-autoscaler-release/metricsforwarder/forwarder"
	"github.com/cloudfoundry/app-autoscaler-release/metricsforwarder/server/auth"
	"github.com/cloudfoundry/app-autoscaler-release/metricsforwarder/server/common"
	"github.com/cloudfoundry/app-autoscaler-release/ratelimiter"
	"github.com/cloudfoundry/app-autoscaler-release/routes"

	"code.cloudfoundry.org/lager"
	"github.com/patrickmn/go-cache"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/http_server"
)

func NewServer(logger lager.Logger, conf *config.Config, policyDB db.PolicyDB, credentials cred_helper.Credentials, allowedMetricCache cache.Cache, httpStatusCollector healthendpoint.HTTPStatusCollector, rateLimiter ratelimiter.Limiter) (ifrit.Runner, error) {
	metricForwarder, err := forwarder.NewMetricForwarder(logger, conf)
	if err != nil {
		logger.Error("failed-to-create-metricforwarder-server", err)
		os.Exit(1)
	}

	mh := NewCustomMetricsHandler(logger, metricForwarder, policyDB, allowedMetricCache)
	authenticator, err := auth.New(logger, credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to add auth middleware: %w", err)
	}
	httpStatusCollectMiddleware := healthendpoint.NewHTTPStatusCollectMiddleware(httpStatusCollector)
	rateLimiterMiddleware := ratelimiter.NewRateLimiterMiddleware("appid", rateLimiter, logger.Session("metricforwarder-ratelimiter-middleware"))

	r := routes.MetricsForwarderRoutes()
	r.Use(rateLimiterMiddleware.CheckRateLimit)
	r.Use(httpStatusCollectMiddleware.Collect)
	r.Use(authenticator.Authenticate)
	r.Get(routes.PostCustomMetricsRouteName).Handler(common.VarsFunc(mh.VerifyCredentialsAndPublishMetrics))

	var addr string
	if os.Getenv("APP_AUTOSCALER_TEST_RUN") == "true" {
		addr = fmt.Sprintf("localhost:%d", conf.Server.Port)
	} else {
		addr = fmt.Sprintf("0.0.0.0:%d", conf.Server.Port)
	}

	runner := http_server.New(addr, r)

	logger.Info("metrics-forwarder-http-server-created", lager.Data{"config": conf})
	return runner, nil
}
