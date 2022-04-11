package sqldb_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	. "changeloglockcleaner/sqldb"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var dbHelper *sqlx.DB

func TestSqldb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sqldb Suite")
}

var _ = BeforeSuite(func() {
	var e error

	dbUrl := os.Getenv("DBURL")
	if dbUrl == "" {
		Fail("environment variable $DBURL is not set")
	}

	database, e := GetConnection(dbUrl)
	if e != nil {
		Fail("failed to parse database connection: " + e.Error())
	}

	dbHelper, e = sqlx.Open(database.DriverName, database.DSN)
	if e != nil {
		Fail("can not connect database: " + e.Error())
	}
})

var _ = AfterSuite(func() {
	if dbHelper != nil {
		dbHelper.Close()
	}

})

func insertLock(id int, locked bool, durationSecond int, lockedby string) (sql.Result, error) {
	var query string
	switch dbHelper.DriverName() {
	case "postgres":
		query = dbHelper.Rebind(fmt.Sprintf("INSERT INTO databasechangeloglock (id,locked,lockgranted,lockedby) VALUES (?,?,now()::timestamp + interval '%d second',?)", durationSecond))
	case "mysql":
		query = dbHelper.Rebind(fmt.Sprintf("INSERT INTO DATABASECHANGELOGLOCK (id,locked,lockgranted,lockedby) VALUES (?,?,date_add(now(),interval %d second) ,?)", durationSecond))
	}
	result, err := dbHelper.Exec(query, id, locked, lockedby)
	return result, err
}

func checkChanglogLockExistenceById(id int) bool {
	var rowCount int
	var query string
	switch dbHelper.DriverName() {
	case "postgres":
		query = dbHelper.Rebind("SELECT COUNT(*) FROM databasechangeloglock WHERE id=?")
	case "mysql":
		query = dbHelper.Rebind("SELECT COUNT(*) FROM DATABASECHANGELOGLOCK WHERE id=?")
	}
	row := dbHelper.QueryRow(query, id)
	_ = row.Scan(&rowCount)
	return rowCount > 0
}

func cleanChanglogLockTable() error {
	var query string
	switch dbHelper.DriverName() {
	case "postgres":
		query = "DELETE FROM databasechangeloglock"
	case "mysql":
		query = "DELETE FROM DATABASECHANGELOGLOCK"
	}
	_, err := dbHelper.Exec(query)
	if err != nil {
		return err
	}
	return nil
}
