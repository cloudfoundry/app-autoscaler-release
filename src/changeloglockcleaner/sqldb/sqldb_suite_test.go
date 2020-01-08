package sqldb_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	. "changeloglockcleaner/sqldb"
	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/jmoiron/sqlx"
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

	database, e := Connection(dbUrl)
	if e != nil {
		Fail("failed to parse database connection: "+ e.Error())
	}

	dbHelper, e =  sqlx.Open(database.DriverName, database.DSN)
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
	query := dbHelper.Rebind(fmt.Sprintf("INSERT INTO databasechangeloglock (id,locked,lockgranted,lockedby) VALUES (?,?,now()::timestamp + interval '%d second',?)", durationSecond))
	result, err := dbHelper.Exec(query, id, locked, lockedby)
	return result, err
}
func checkChanglogLockExistenceById(id int) bool {
	var rowCount int
	query := dbHelper.Rebind("SELECT COUNT(*) FROM databasechangeloglock WHERE id=?")
	row := dbHelper.QueryRow(query, id)
	row.Scan(&rowCount)
	return rowCount > 0
}
func cleanChanglogLockTable() error {
	_, err := dbHelper.Exec("DELETE FROM databasechangeloglock")
	if err != nil {
		return err
	}
	return nil
}
