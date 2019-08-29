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
)

var dbHelper *sql.DB

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

	dbHelper, e = sql.Open(PostgresDriverName, dbUrl)
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
	query := fmt.Sprintf("INSERT INTO databasechangeloglock (id,locked,lockgranted,lockedby) VALUES ($1,$2,now()::timestamp + interval '%d second',$3)", durationSecond)
	result, err := dbHelper.Exec(query, id, locked, lockedby)
	return result, err
}
func checkChanglogLockExistenceById(id int) bool {
	var rowCount int
	query := "SELECT COUNT(*) FROM databasechangeloglock WHERE id=$1"
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
