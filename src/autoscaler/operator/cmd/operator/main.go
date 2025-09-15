package main

import (
	"os"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/cf"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/configutil"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/db"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/db/sqldb"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/healthendpoint"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/operator"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/operator/config"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/startup"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/sync"
	"github.com/google/uuid"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager/v3"
	"github.com/prometheus/client_golang/prometheus"
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

	logger := startup.InitLogger(&conf.Logging, "operator")
	prClock := clock.NewClock()

	appMetricsDB, err := sqldb.NewAppMetricSQLDB(conf.Db[db.AppMetricsDb], logger.Session("appmetrics-db"))
	startup.ExitOnError(err, logger, "failed to connect appmetrics db", lager.Data{"dbConfig": conf.Db[db.AppMetricsDb]})
	defer appMetricsDB.Close()

	scalingEngineDB, err := sqldb.NewScalingEngineSQLDB(conf.Db[db.ScalingEngineDb], logger.Session("scalingengine-db"))
	startup.ExitOnError(err, logger, "failed to connect scalingengine db", lager.Data{"dbConfig": conf.Db[db.ScalingEngineDb]})
	defer scalingEngineDB.Close()

	cfClient := cf.NewCFClient(&conf.CF, logger.Session("cf"), prClock)
	err = cfClient.Login()
	startup.ExitOnError(err, logger, "failed to login cloud foundry", lager.Data{"API": conf.CF.API})

	policyDb, err := sqldb.NewPolicySQLDB(conf.Db[db.PolicyDb], logger.Session("policy-db"))
	startup.ExitOnError(err, logger, "failed to connect policy db", lager.Data{"dbConfig": conf.Db[db.PolicyDb]})
	defer policyDb.Close()

	scalingEngineHttpclient, err := helpers.CreateHTTPSClient(&conf.ScalingEngine.TLSClientCerts, helpers.DefaultClientConfig(), logger.Session("scaling_client"))
	startup.ExitOnError(err, logger, "failed to create http client for scalingengine", lager.Data{"scalingengineTLS": conf.ScalingEngine.TLSClientCerts})
	
	schedulerHttpclient, err := helpers.CreateHTTPSClient(&conf.Scheduler.TLSClientCerts, helpers.DefaultClientConfig(), logger.Session("scheduler_client"))
	startup.ExitOnError(err, logger, "failed to create http client for scheduler", lager.Data{"schedulerTLS": conf.Scheduler.TLSClientCerts})

	loggerSessionName := "appmetrics-dbpruner"
	appMetricsDBPruner := operator.NewAppMetricsDbPruner(appMetricsDB, conf.AppMetricsDb.CutoffDuration, prClock, logger.Session(loggerSessionName))
	appMetricsDBOperatorRunner := operator.NewOperatorRunner(appMetricsDBPruner, conf.AppMetricsDb.RefreshInterval, prClock, logger.Session(loggerSessionName))

	loggerSessionName = "scalingengine-dbpruner"
	scalingEngineDBPruner := operator.NewScalingEngineDbPruner(scalingEngineDB, conf.ScalingEngineDb.CutoffDuration, prClock, logger.Session(loggerSessionName))
	scalingEngineDBOperatorRunner := operator.NewOperatorRunner(scalingEngineDBPruner, conf.ScalingEngineDb.RefreshInterval, prClock, logger.Session(loggerSessionName))
	loggerSessionName = "scalingengine-sync"
	scalingEngineSync := operator.NewScheduleSynchronizer(scalingEngineHttpclient, conf.ScalingEngine.URL, prClock, logger.Session(loggerSessionName))
	scalingEngineSyncRunner := operator.NewOperatorRunner(scalingEngineSync, conf.ScalingEngine.SyncInterval, prClock, logger.Session(loggerSessionName))

	loggerSessionName = "scheduler-sync"
	schedulerSync := operator.NewScheduleSynchronizer(schedulerHttpclient, conf.Scheduler.URL, prClock, logger.Session(loggerSessionName))
	schedulerSyncRunner := operator.NewOperatorRunner(schedulerSync, conf.Scheduler.SyncInterval, prClock, logger.Session(loggerSessionName))

	loggerSessionName = "application-sync"
	applicationSync := operator.NewApplicationSynchronizer(cfClient.GetCtxClient(), policyDb, logger.Session(loggerSessionName))
	applicationSyncRunner := operator.NewOperatorRunner(applicationSync, conf.AppSyncer.SyncInterval, prClock, logger.Session(loggerSessionName))

	members := grouper.Members{
		{Name: "appmetrics-dbpruner", Runner: appMetricsDBOperatorRunner},
		{Name: "scalingEngine-dbpruner", Runner: scalingEngineDBOperatorRunner},
		{Name: "scalingEngine-sync", Runner: scalingEngineSyncRunner},
		{Name: "scheduler-sync", Runner: schedulerSyncRunner},
		{Name: "application-sync", Runner: applicationSyncRunner},
	}

	guid := uuid.NewString()
	const lockTableName = "operator_lock"
	var lockDB db.LockDB
	lockDB, err = sqldb.NewLockSQLDB(conf.Db[db.LockDb], lockTableName, logger.Session("lock-db"))
	startup.ExitOnError(err, logger, "failed-to-connect-lock-database", lager.Data{"dbConfig": conf.Db[db.LockDb]})
	defer lockDB.Close()
	prdl := sync.NewDatabaseLock(logger)
	dbLockMaintainer := prdl.InitDBLockRunner(conf.DBLock.LockRetryInterval, conf.DBLock.LockTTL, guid, lockDB, func() {}, func() {
		os.Exit(1)
	})

	members = append(
		grouper.Members{{Name: "db-lock-maintainer", Runner: dbLockMaintainer}},
		members...,
	)
	gatherer := createPrometheusRegistry(policyDb, appMetricsDB, scalingEngineDB, logger)
	healthRouter, err := healthendpoint.NewHealthRouter(conf.Health, []healthendpoint.Checker{}, logger, gatherer, time.Now)
	startup.ExitOnError(err, logger, "failed to create health router")

	healthServer, err := helpers.NewHTTPServer(logger, conf.Health.ServerConfig, healthRouter)
	startup.ExitOnError(err, logger, "failed to create health server")

	members = append(
		grouper.Members{{Name: "health_server", Runner: healthServer}},
		members...,
	)

	err = startup.StartServices(logger, members)
	if err != nil {
		os.Exit(1)
	}
}

func createPrometheusRegistry(policyDB db.PolicyDB, appMetricsDB db.AppMetricDB, scalingEngineDB db.ScalingEngineDB, logger lager.Logger) *prometheus.Registry {
	promRegistry := prometheus.NewRegistry()
	healthendpoint.RegisterCollectors(promRegistry, []prometheus.Collector{
		healthendpoint.NewDatabaseStatusCollector("autoscaler", "operator", "policyDB", policyDB),
		healthendpoint.NewDatabaseStatusCollector("autoscaler", "operator", "appMetricsDB", appMetricsDB),
		healthendpoint.NewDatabaseStatusCollector("autoscaler", "operator", "scalingEngineDB", scalingEngineDB),
	}, true, logger.Session("operator-prometheus"))
	return promRegistry
}

