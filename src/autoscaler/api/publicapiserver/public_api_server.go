package publicapiserver

import (
	"fmt"
	"net/http"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/cred_helper"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/api"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/api/apis/scalinghistory"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/api/config"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/cf"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/db"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/healthendpoint"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/ratelimiter"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/routes"

	"code.cloudfoundry.org/lager/v3"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tedsuo/ifrit"
)

type VarsFunc func(w http.ResponseWriter, r *http.Request, vars map[string]string)

func (vh VarsFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vh(w, r, vars)
}

func NewPublicApiServer(logger lager.Logger, conf *config.Config, policyDB db.PolicyDB,
	bindingDB db.BindingDB, credentials cred_helper.Credentials, checkBindingFunc api.CheckBindingFunc,
	cfclient cf.CFClient, httpStatusCollector healthendpoint.HTTPStatusCollector,
	rateLimiter ratelimiter.Limiter) (ifrit.Runner, error) {
	pah := NewPublicApiHandler(logger, conf, policyDB, bindingDB, credentials)

	scalingHistoryHandler, err := newScalingHistoryHandler(logger, conf)
	if err != nil {
		return nil, err
	}

	mw := NewMiddleware(logger, cfclient, checkBindingFunc, conf.APIClientId)
	rateLimiterMiddleware := ratelimiter.NewRateLimiterMiddleware("appId", rateLimiter, logger.Session("api-ratelimiter-middleware"))
	httpStatusCollectMiddleware := healthendpoint.NewHTTPStatusCollectMiddleware(httpStatusCollector)

	r := routes.ApiOpenRoutes()
	r.Use(otelmux.Middleware("apiserver"))
	r.Use(httpStatusCollectMiddleware.Collect)

	r.Get(routes.PublicApiInfoRouteName).Handler(VarsFunc(pah.GetApiInfo))
	r.Get(routes.PublicApiHealthRouteName).Handler(VarsFunc(pah.GetHealth))

	rp := routes.ApiRoutes()
	rp.Use(rateLimiterMiddleware.CheckRateLimit)
	rp.Use(mw.HasClientToken)
	rp.Use(mw.Oauth)
	rp.Use(httpStatusCollectMiddleware.Collect)

	rp.Get(routes.PublicApiScalingHistoryRouteName).Handler(scalingHistoryHandler)
	rp.Get(routes.PublicApiAggregatedMetricsHistoryRouteName).Handler(VarsFunc(pah.GetAggregatedMetricsHistories))

	rpolicy := routes.ApiPolicyRoutes()
	rpolicy.Use(rateLimiterMiddleware.CheckRateLimit)
	rpolicy.Use(mw.HasClientToken)
	rpolicy.Use(mw.Oauth)
	rpolicy.Use(mw.CheckServiceBinding)

	rpolicy.Use(httpStatusCollectMiddleware.Collect)
	rpolicy.Get(routes.PublicApiGetPolicyRouteName).Handler(VarsFunc(pah.GetScalingPolicy))
	rpolicy.Get(routes.PublicApiAttachPolicyRouteName).Handler(VarsFunc(pah.AttachScalingPolicy))
	rpolicy.Get(routes.PublicApiDetachPolicyRouteName).Handler(VarsFunc(pah.DetachScalingPolicy))

	rcredential := routes.ApiCredentialRoutes()
	rcredential.Use(rateLimiterMiddleware.CheckRateLimit)

	rcredential.Use(httpStatusCollectMiddleware.Collect)
	rcredential.Use(mw.HasClientToken)
	rcredential.Use(mw.Oauth)

	healthRouter, err := createHealthRouter(logger, conf, policyDB, bindingDB, httpStatusCollector)
	if err != nil {
		return nil, fmt.Errorf("failed to create health router: %w", err)
	}

	mainRouter := setupMainRouter(r, healthRouter)

	return helpers.NewHTTPServer(logger, conf.PublicApiServer, mainRouter)
}
func setupMainRouter(r *mux.Router, healthRouter *mux.Router) *mux.Router {
	mainRouter := mux.NewRouter()
	mainRouter.PathPrefix("/v1").Handler(r)
	mainRouter.PathPrefix("/health").Handler(healthRouter)
	mainRouter.PathPrefix("/").Handler(healthRouter)
	return mainRouter
}

func createPrometheusRegistry(policyDB db.PolicyDB, bindingDB db.BindingDB, httpStatusCollector healthendpoint.HTTPStatusCollector, logger lager.Logger) *prometheus.Registry {
	promRegistry := prometheus.NewRegistry()
	healthendpoint.RegisterCollectors(promRegistry,
		[]prometheus.Collector{
			healthendpoint.NewDatabaseStatusCollector("autoscaler", "golangapiserver", "policyDB", policyDB),
			healthendpoint.NewDatabaseStatusCollector("autoscaler", "golangapiserver", "bindingDB", bindingDB),
			httpStatusCollector,
		},
		true, logger.Session("golangapiserver-prometheus"))
	return promRegistry
}

func createHealthRouter(logger lager.Logger, conf *config.Config, policyDB db.PolicyDB, bindingDB db.BindingDB, httpStatusCollector healthendpoint.HTTPStatusCollector) (*mux.Router, error) {
	checkers := []healthendpoint.Checker{}
	gatherer := createPrometheusRegistry(policyDB, bindingDB, httpStatusCollector, logger)
	healthRouter, err := healthendpoint.NewHealthRouter(conf.Health, checkers, logger.Session("health-server"), gatherer, time.Now)
	if err != nil {
		return nil, fmt.Errorf("failed to create health router: %w", err)
	}

	logger.Debug("Successfully created health server")
	return healthRouter, nil
}

func newScalingHistoryHandler(logger lager.Logger, conf *config.Config) (http.Handler, error) {
	scalingHistoryHandler, err := NewScalingHistoryHandler(logger, conf)
	if err != nil {
		return nil, fmt.Errorf("error creating scaling history handler: %w", err)
	}
	scalingHistoryServer, err := scalinghistory.NewServer(scalingHistoryHandler, scalingHistoryHandler)
	if err != nil {
		return nil, fmt.Errorf("error creating ogen scaling history server: %w", err)
	}
	return scalingHistoryServer, nil
}
