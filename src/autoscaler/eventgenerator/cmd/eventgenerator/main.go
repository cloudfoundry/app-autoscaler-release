package main

import (
	"os"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/configutil"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/db"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/db/sqldb"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/eventgenerator/aggregator"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/eventgenerator/config"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/eventgenerator/generator"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/eventgenerator/metric"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/eventgenerator/server"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/healthendpoint"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers/auth"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/startup"
	"github.com/prometheus/client_golang/prometheus"
	circuit "github.com/rubyist/circuitbreaker"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager/v3"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
)

type configLoader struct{}

func (c *configLoader) LoadConfig(path string, vcapConfigReader configutil.VCAPConfigurationReader) (*config.Config, error) {
	return config.LoadConfig(path, vcapConfigReader)
}

func main() {
	path := startup.ParseFlags()
	
	vcapConfiguration, _ := startup.LoadVCAPConfiguration()
	
	conf, err := startup.LoadAndValidateConfig(path, vcapConfiguration, &configLoader{})
	if err != nil {
		os.Exit(1)
	}

	startup.SetupEnvironment()

	logger := startup.InitLogger(&conf.Logging, "eventgenerator")

	egClock := clock.NewClock()

	appMetricDB, err := sqldb.NewAppMetricSQLDB(conf.Db[db.AppMetricsDb], logger.Session("appMetric-db"))
	startup.ExitOnError(err, logger, "failed to connect app-metric database", lager.Data{"dbConfig": conf.Db[db.AppMetricsDb]})
	defer func() { _ = appMetricDB.Close() }()

	policyDb := sqldb.CreatePolicyDb(conf.Db[db.PolicyDb], logger)
	defer func() { _ = policyDb.Close() }()

	httpStatusCollector := healthendpoint.NewHTTPStatusCollector("autoscaler", "eventgenerator")
	promRegistry := prometheus.NewRegistry()
	healthendpoint.RegisterCollectors(promRegistry, []prometheus.Collector{
		healthendpoint.NewDatabaseStatusCollector("autoscaler", "eventgenerator", "appMetricDB", appMetricDB),
		healthendpoint.NewDatabaseStatusCollector("autoscaler", "eventgenerator", "policyDB", policyDb),
		httpStatusCollector,
	}, true, logger.Session("eventgenerator-prometheus"))

	//appManager := aggregator.NewAppManager(logger, egClock, conf.Aggregator.PolicyPollerInterval, conf.Pool.TotalInstances, conf.Pool.InstanceIndex, conf.Aggregator.MetricCacheSizePerApp, policyDb, appMetricDB)
	appManager := aggregator.NewAppManager(logger, egClock, *conf.Aggregator, *conf.Pool, policyDb, appMetricDB)

	triggersChan := make(chan []*models.Trigger, conf.Evaluator.TriggerArrayChannelSize)

	evaluationManager, err := generator.NewAppEvaluationManager(logger, conf.Evaluator.EvaluationManagerInterval, egClock, triggersChan, appManager.GetPolicies, *conf.CircuitBreaker)
	startup.ExitOnError(err, logger, "failed to create Evaluation Manager")

	evaluators, err := createEvaluators(logger, conf, triggersChan, appManager.QueryAppMetrics, evaluationManager.GetBreaker, evaluationManager.SetCoolDownExpired)
	startup.ExitOnError(err, logger, "failed to create Evaluators")

	appMonitorsChan := make(chan *models.AppMonitor, conf.Aggregator.AppMonitorChannelSize)
	appMetricChan := make(chan *models.AppMetric, conf.Aggregator.AppMetricChannelSize)

	fetcherFactory := metric.NewLogCacheFetcherFactory(metric.StandardLogCacheFetcherCreator)
	metricFetcher, err := fetcherFactory.CreateFetcher(logger, *conf)
	startup.ExitOnError(err, logger, "failed to create metric fetcher")

	metricPollers, err := createMetricPollers(logger, conf, appMonitorsChan, appMetricChan, metricFetcher)
	startup.ExitOnError(err, logger, "failed to create MetricPoller")
	
	anAggregator, err := aggregator.NewAggregator(logger, egClock, conf.Aggregator.AggregatorExecuteInterval, conf.Aggregator.SaveInterval, appMonitorsChan, appManager.GetPolicies, appManager.SaveMetricToCache, conf.DefaultStatWindowSecs, appMetricChan, appMetricDB)
	startup.ExitOnError(err, logger, "failed to create Aggregator")

	eventGenerator := ifrit.RunFunc(runFunc(appManager, evaluators, evaluationManager, metricPollers, anAggregator))

	eventgeneratorServer := server.NewServer(
		logger.Session("http_server"), conf, appMetricDB,
		policyDb, appManager.QueryAppMetrics, httpStatusCollector)

	mtlsServer, err := eventgeneratorServer.CreateMtlsServer()
	startup.ExitOnError(err, logger, "failed to create http server")

	healthServer, err := eventgeneratorServer.CreateHealthServer()
	startup.ExitOnError(err, logger, "failed to create health server")

	xm := auth.NewXfccAuthMiddleware(logger, conf.CFServer.XFCC)
	cfServer, err := eventgeneratorServer.CreateCFServer(xm)
	startup.ExitOnError(err, logger, "failed to create cf server")

	members := grouper.Members{
		{Name: "eventGenerator", Runner: eventGenerator},
		{Name: "https_server", Runner: mtlsServer},
		{Name: "health_server", Runner: healthServer},
		{Name: "cf_server", Runner: cfServer},
	}

	err = startup.StartServices(logger, members)
	if err != nil {
		os.Exit(1)
	}
}

func runFunc(appManager *aggregator.AppManager, evaluators []*generator.Evaluator, evaluationManager *generator.AppEvaluationManager, metricPollers []*aggregator.MetricPoller, anAggregator *aggregator.Aggregator) func(signals <-chan os.Signal, ready chan<- struct{}) error {
	return func(signals <-chan os.Signal, ready chan<- struct{}) error {
		appManager.Start()

		for _, evaluator := range evaluators {
			evaluator.Start()
		}
		evaluationManager.Start()

		for _, metricPoller := range metricPollers {
			metricPoller.Start()
		}
		anAggregator.Start()

		close(ready)

		<-signals
		anAggregator.Stop()
		evaluationManager.Stop()
		appManager.Stop()

		return nil
	}
}

func createEvaluators(logger lager.Logger, conf *config.Config, triggersChan chan []*models.Trigger, queryMetrics aggregator.QueryAppMetricsFunc, getBreaker func(string) *circuit.Breaker, setCoolDownExpired func(string, int64)) ([]*generator.Evaluator, error) {
	count := conf.Evaluator.EvaluatorCount

	seClient, err := helpers.CreateHTTPSClient(&conf.ScalingEngine.TLSClientCerts, helpers.DefaultClientConfig(), logger.Session("scaling_client"))
	if err != nil {
		return nil, err
	}

	evaluators := make([]*generator.Evaluator, count)
	for i := range evaluators {
		evaluators[i] = generator.NewEvaluator(logger, seClient, conf.ScalingEngine.ScalingEngineURL, triggersChan,
			conf.DefaultBreachDurationSecs, queryMetrics, getBreaker, setCoolDownExpired)
	}

	return evaluators, nil
}

func createMetricPollers(logger lager.Logger, conf *config.Config, appMonitorsChan chan *models.AppMonitor, appMetricChan chan *models.AppMetric, metricClient metric.Fetcher) ([]*aggregator.MetricPoller, error) {
	pollers := make([]*aggregator.MetricPoller, conf.Aggregator.MetricPollerCount)
	for i := range pollers {
		pollers[i] = aggregator.NewMetricPoller(logger, metricClient, appMonitorsChan, appMetricChan)
	}
	return pollers, nil
}

