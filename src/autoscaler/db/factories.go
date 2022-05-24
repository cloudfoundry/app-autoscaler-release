package db

import (
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/db/sqldb"
	"code.cloudfoundry.org/lager"
	"os"
)

func CreatePolicyDb(dbConf DatabaseConfig, logger lager.Logger) *sqldb.PolicySQLDB {
	policyDB, err := sqldb.NewPolicySQLDB(dbConf, logger.Session("policy-db"))
	if err != nil {
		logger.Fatal("Failed To connect to policyDB", err, lager.Data{"dbConfig": dbConf.Db[db.PolicyDb]})
		os.Exit(1)
	}
	return policyDB
}
