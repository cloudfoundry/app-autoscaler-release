package server

import (
	"fmt"
	"net/http"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/db"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/eventgenerator/aggregator"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/eventgenerator/config"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/healthendpoint"
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

func NewServer(logger lager.Logger, conf *config.Config, appMetricDB db.AppMetricDB, policyDb db.PolicyDB, queryAppMetric aggregator.QueryAppMetricsFunc, httpStatusCollector healthendpoint.HTTPStatusCollector) (ifrit.Runner, error) {
	eh := NewEventGenHandler(logger, queryAppMetric)
	httpStatusCollectMiddleware := healthendpoint.NewHTTPStatusCollectMiddleware(httpStatusCollector)
	r := routes.EventGeneratorRoutes()
	r.Use(otelmux.Middleware("eventgenerator"))
	r.Use(httpStatusCollectMiddleware.Collect)
	r.Get(routes.GetAggregatedMetricHistoriesRouteName).Handler(VarsFunc(eh.GetAggregatedMetricHistories))

	healthRouter, err := createHealthRouter(appMetricDB, policyDb, logger, conf, httpStatusCollector)
	if err != nil {
		return nil, fmt.Errorf("failed to create health router: %w", err)
	}

	mainRouter := setupMainRouter(r, healthRouter)
	return helpers.NewHTTPServer(logger, serverConfigFrom(conf), mainRouter)
}

func serverConfigFrom(conf *config.Config) helpers.ServerConfig {
	return helpers.ServerConfig{
		Port: conf.Server.Port,
		TLS:  conf.Server.TLS,
	}
}

func createHealthRouter(appMetricDB db.AppMetricDB, policyDb db.PolicyDB, logger lager.Logger, conf *config.Config, httpStatusCollector healthendpoint.HTTPStatusCollector) (*mux.Router, error) {
	checkers := []healthendpoint.Checker{}
	gatherer := CreatePrometheusRegistry(appMetricDB, policyDb, httpStatusCollector, logger)
	healthRouter, err := healthendpoint.NewHealthRouter(conf.Health, checkers, logger.Session("health-server"), gatherer, time.Now)
	if err != nil {
		return nil, fmt.Errorf("failed to create health router: %w", err)
	}

	return healthRouter, nil
}

func CreatePrometheusRegistry(appMetricDB db.AppMetricDB, policyDb db.PolicyDB, httpStatusCollector healthendpoint.HTTPStatusCollector, logger lager.Logger) *prometheus.Registry {
	promRegistry := prometheus.NewRegistry()
	healthendpoint.RegisterCollectors(promRegistry, []prometheus.Collector{
		healthendpoint.NewDatabaseStatusCollector("autoscaler", "eventgenerator", "appMetricDB", appMetricDB),
		healthendpoint.NewDatabaseStatusCollector("autoscaler", "eventgenerator", "policyDB", policyDb),
		httpStatusCollector,
	}, true, logger.Session("eventgenerator-prometheus"))
	return promRegistry
}

func setupMainRouter(egRouter, healthRouter *mux.Router) *mux.Router {
	mainRouter := mux.NewRouter()
	mainRouter.PathPrefix("/v1").Handler(egRouter)
	mainRouter.PathPrefix("/health").Handler(healthRouter)
	mainRouter.PathPrefix("/").Handler(healthRouter)
	return mainRouter
}
