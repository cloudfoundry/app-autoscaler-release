package main

import (
	"os"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/cf"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/configutil"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/db"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/db/sqldb"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers/auth"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/scalingengine"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/scalingengine/config"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/scalingengine/schedule"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/scalingengine/server"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/startup"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager/v3"
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

	logger := startup.InitLogger(&conf.Logging, "scalingengine")

	eClock := clock.NewClock()
	cfClient := cf.NewCFClient(&conf.CF, logger.Session("cf"), eClock)
	err = cfClient.Login()
	startup.ExitOnError(err, logger, "failed to login cloud foundry", lager.Data{"API": conf.CF.API})

	policyDb, err := sqldb.NewPolicySQLDB(conf.Db[db.PolicyDb], logger.Session("policy-db"))
	startup.ExitOnError(err, logger, "failed to connect policy db", lager.Data{"dbConfig": conf.Db[db.PolicyDb]})
	defer policyDb.Close()

	scalingEngineDB, err := sqldb.NewScalingEngineSQLDB(conf.Db[db.ScalingEngineDb], logger.Session("scalingengine-db"))
	startup.ExitOnError(err, logger, "failed to connect scalingengine database", lager.Data{"dbConfig": conf.Db[db.ScalingEngineDb]})
	defer func() { _ = scalingEngineDB.Close() }()

	schedulerDB, err := sqldb.NewSchedulerSQLDB(conf.Db[db.SchedulerDb], logger.Session("scheduler-db"))
	startup.ExitOnError(err, logger, "failed to connect scheduler database", lager.Data{"dbConfig": conf.Db[db.SchedulerDb]})
	defer func() { _ = schedulerDB.Close() }()

	scalingEngine := scalingengine.NewScalingEngine(logger, cfClient, policyDb, scalingEngineDB, eClock, conf.DefaultCoolDownSecs, conf.LockSize)
	synchronizer := schedule.NewActiveScheduleSychronizer(logger, schedulerDB, scalingEngineDB, scalingEngine)

	server := server.NewServer(logger.Session("http-server"), conf, policyDb, scalingEngineDB, schedulerDB, scalingEngine, synchronizer)
	httpServer, err := server.CreateMtlsServer()
	startup.ExitOnError(err, logger, "failed to create http server")

	healthServer, err := server.CreateHealthServer()
	startup.ExitOnError(err, logger, "failed to create health server")

	xm := auth.NewXfccAuthMiddleware(logger, conf.CFServer.XFCC)
	cfServer, err := server.CreateCFServer(xm)
	startup.ExitOnError(err, logger, "failed to create cf server")

	members := grouper.Members{
		{Name: "http_server", Runner: httpServer},
		{Name: "health_server", Runner: healthServer},
		{Name: "cf_server", Runner: cfServer},
	}

	err = startup.StartServices(logger, members)
	if err != nil {
		os.Exit(1)
	}
}

