package main

import (
	"io"

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
	"github.com/prometheus/client_golang/prometheus"
	circuit "github.com/rubyist/circuitbreaker"

	"flag"
	"fmt"
	"os"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager/v3"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"
)

func main() {
	var path string
	flag.StringVar(&path, "c", "", "config file")
	flag.Parse()
	if path == "" {
		_, _ = fmt.Fprintln(os.Stdout, "missing config file\nUsage:use '-c' option to specify the config file path")
		os.Exit(1)
	}
	conf, err := loadConfig(path)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "%s\n", err.Error())
		os.Exit(1)
	}

	helpers.SetupOpenTelemetry()

	logger := helpers.InitLoggerFromConfig(&conf.Logging, "eventgenerator")
	egClock := clock.NewClock()

	appMetricDB, err := sqldb.NewAppMetricSQLDB(conf.DB.AppMetricDB, logger.Session("appMetric-db"))
	if err != nil {
		logger.Error("failed to connect app-metric database", err, lager.Data{"dbConfig": conf.DB.AppMetricDB})
		os.Exit(1)
	}
	defer func() { _ = appMetricDB.Close() }()

	policyDb := sqldb.CreatePolicyDb(conf.DB.PolicyDB, logger)
	defer func() { _ = policyDb.Close() }()

	httpStatusCollector := healthendpoint.NewHTTPStatusCollector("autoscaler", "eventgenerator")
	promRegistry := prometheus.NewRegistry()
	healthendpoint.RegisterCollectors(promRegistry, []prometheus.Collector{
		healthendpoint.NewDatabaseStatusCollector("autoscaler", "eventgenerator", "appMetricDB", appMetricDB),
		healthendpoint.NewDatabaseStatusCollector("autoscaler", "eventgenerator", "policyDB", policyDb),
		httpStatusCollector,
	}, true, logger.Session("eventgenerator-prometheus"))

	appManager := aggregator.NewAppManager(logger, egClock, conf.Aggregator.PolicyPollerInterval, len(conf.Server.NodeAddrs), conf.Server.NodeIndex, conf.Aggregator.MetricCacheSizePerApp, policyDb, appMetricDB)

	triggersChan := make(chan []*models.Trigger, conf.Evaluator.TriggerArrayChannelSize)

	evaluationManager, err := generator.NewAppEvaluationManager(logger, conf.Evaluator.EvaluationManagerInterval, egClock, triggersChan, appManager.GetPolicies, conf.CircuitBreaker)
	if err != nil {
		logger.Error("failed to create Evaluation Manager", err)
		os.Exit(1)
	}

	evaluators, err := createEvaluators(logger, conf, triggersChan, appManager.QueryAppMetrics, evaluationManager.GetBreaker, evaluationManager.SetCoolDownExpired)
	if err != nil {
		logger.Error("failed to create Evaluators", err)
		os.Exit(1)
	}

	appMonitorsChan := make(chan *models.AppMonitor, conf.Aggregator.AppMonitorChannelSize)
	appMetricChan := make(chan *models.AppMetric, conf.Aggregator.AppMetricChannelSize)

	fetcherFactory := metric.NewLogCacheFetcherFactory(metric.StandardLogCacheFetcherCreator)
	metricFetcher, err := fetcherFactory.CreateFetcher(logger, *conf)
	if err != nil {
		logger.Error("failed to create metric fetcher", err)
		os.Exit(1)
	}

	metricPollers, err := createMetricPollers(logger, conf, appMonitorsChan, appMetricChan, metricFetcher)
	if err != nil {
		logger.Error("failed to create MetricPoller", err)
		os.Exit(1)
	}
	anAggregator, err := aggregator.NewAggregator(logger, egClock, conf.Aggregator.AggregatorExecuteInterval, conf.Aggregator.SaveInterval, appMonitorsChan, appManager.GetPolicies, appManager.SaveMetricToCache, conf.DefaultStatWindowSecs, appMetricChan, appMetricDB)
	if err != nil {
		logger.Error("failed to create Aggregator", err)
		os.Exit(1)
	}

	eventGenerator := ifrit.RunFunc(runFunc(appManager, evaluators, evaluationManager, metricPollers, anAggregator))

	httpServer := server.NewServer(logger.Session("http_server"), conf, appMetricDB, policyDb, appManager.QueryAppMetrics, httpStatusCollector)

	mtlsServer, err := httpServer.CreateMtlsServer()
	if err != nil {
		logger.Error("failed to create http server", err)
		os.Exit(1)
	}

	healthServer, err := httpServer.CreateHealthServer()
	if err != nil {
		logger.Error("failed to create health server", err)
		os.Exit(1)
	}

	xm := auth.NewXfccAuthMiddleware(logger, conf.CFServer.XFCC)
	cfServer, err := httpServer.CreateCFServer(xm)
	if err != nil {
		logger.Error("failed to create cf server", err)
		os.Exit(1)
	}

	members := grouper.Members{
		{"eventGenerator", eventGenerator},
		{"https_server", mtlsServer},
		{"health_server", healthServer},
		{"cf_server", cfServer},
	}

	monitor := ifrit.Invoke(sigmon.New(grouper.NewOrdered(os.Interrupt, members)))
	logger.Info("started")
	err = <-monitor.Wait()
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}
	logger.Info("exited")
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
func loadConfig(path string) (*config.Config, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file %q: %w", path, err)
	}

	configFileBytes, err := io.ReadAll(configFile)
	defer func() { _ = configFile.Close() }()

	if err != nil {
		return nil, fmt.Errorf("failed to read data from config file %q: %w", path, err)
	}

	conf, err := config.LoadConfig(configFileBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file %q: %w", path, err)
	}

	err = conf.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate configuration: %w", err)
	}
	return conf, nil
}

func createEvaluators(logger lager.Logger, conf *config.Config, triggersChan chan []*models.Trigger, queryMetrics aggregator.QueryAppMetricsFunc, getBreaker func(string) *circuit.Breaker, setCoolDownExpired func(string, int64)) ([]*generator.Evaluator, error) {
	count := conf.Evaluator.EvaluatorCount

	seClient, err := helpers.CreateHTTPSClient(&conf.ScalingEngine.TLSClientCerts, helpers.DefaultClientConfig(), logger.Session("scaling_client"))
	if err != nil {
		logger.Error("failed to create http client for ScalingEngine", err, lager.Data{"scalingengineTLS": conf.ScalingEngine.TLSClientCerts})
		os.Exit(1)
	}

	evaluators := make([]*generator.Evaluator, count)
	for i := 0; i < count; i++ {
		evaluators[i] = generator.NewEvaluator(logger, seClient, conf.ScalingEngine.ScalingEngineURL, triggersChan,
			conf.DefaultBreachDurationSecs, queryMetrics, getBreaker, setCoolDownExpired)
	}

	return evaluators, nil
}

func createMetricPollers(logger lager.Logger, conf *config.Config, appMonitorsChan chan *models.AppMonitor, appMetricChan chan *models.AppMetric, metricClient metric.Fetcher) ([]*aggregator.MetricPoller, error) {
	pollers := make([]*aggregator.MetricPoller, conf.Aggregator.MetricPollerCount)
	for i := 0; i < len(pollers); i++ {
		pollers[i] = aggregator.NewMetricPoller(logger, metricClient, appMonitorsChan, appMetricChan)
	}
	return pollers, nil
}
