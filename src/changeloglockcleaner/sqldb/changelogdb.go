package sqldb

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

const PostgresDriverName = "pgx"
const MysqlDriverName = "mysql"
const postgresDbURLPattern = `^(postgres|postgresql):\/\/(.+):(.+)@([\da-zA-Z\.-]+)(:[\d]{4,5})?\/(.+)`
const mysqlDbURLPattern = `(.+):(.+)@tcp\(([\da-zA-Z\.-]+)(:[\d]{4,5})?\)\/(.+)`

type ChangelogSQLDB struct {
	sqldb *sqlx.DB
}

func NewChangelogSQLDB(dbUrl string) (*ChangelogSQLDB, error) {
	log.SetOutput(os.Stdout)
	database, err := GetConnection(dbUrl)
	if err != nil {
		return nil, err
	}

	sqldb, err := sqlx.Open(database.DriverName, database.DataSourceName)
	if err != nil {
		return nil, err
	}

	err = sqldb.Ping()
	if err != nil {
		sqldb.Close()
		dbUrl = redactDbCreds(dbUrl)
		if dbUrl != "" {
			log.Printf("failed-to-connection-to-database, dburl:%s,  err:%s\n", dbUrl, err)
		}
		return nil, err
	}

	return &ChangelogSQLDB{
		sqldb: sqldb,
	}, nil
}

func (cdb *ChangelogSQLDB) Close() error {
	err := cdb.sqldb.Close()
	if err != nil {
		log.Printf("failed-to-close-connection, err:%s\n", err)
		return err
	}
	return nil
}

func (cdb *ChangelogSQLDB) DeleteExpiredLock(timeoutInSecond int) error {
	switch cdb.sqldb.DriverName() {
	case "pgx":
		query := fmt.Sprintf(`DO $$
	BEGIN
		IF EXISTS
			( SELECT 1
			  FROM   information_schema.tables
			  WHERE  table_schema = 'public'
			  AND    table_name = 'databasechangeloglock'
			)
		THEN
			DELETE FROM databasechangeloglock WHERE EXTRACT(epoch FROM (now()::timestamp - lockgranted))::int > %d;
		END IF ;
	END
   $$ ;
	`, timeoutInSecond)
		_, err := cdb.sqldb.Exec(query)
		if err != nil {
			log.Printf("failed-to-delete-application-details, query:%s, err:%s\n", query, err)
		}
		return err
	case "mysql":
		var rowCount int
		err := cdb.sqldb.QueryRow("SELECT 1 FROM  information_schema.tables WHERE  table_schema = 'autoscaler' AND table_name = 'DATABASECHANGELOGLOCK'").Scan(&rowCount)
		if err == sql.ErrNoRows {
			log.Printf("table databasechangeloglock does not exist, err:%s\n", err)
			return nil
		} else if err != nil {
			log.Printf("failed to query table from database, err:%s\n", err)
			return err
		}
		if rowCount > 0 {
			_, err = cdb.sqldb.Exec(fmt.Sprintf("DELETE FROM DATABASECHANGELOGLOCK WHERE TIMESTAMPDIFF(SECOND, lockgranted, NOW()) > %d;", timeoutInSecond))
			if err != nil {
				log.Printf("failed-to-delete-application-details, err:%s\n", err)
				return err
			}
		}
	}
	return nil
}

func redactDbCreds(dbUrl string) string {
	var redactUrl string
	var urlCredMatcher *regexp.Regexp
	if strings.Contains(dbUrl, "postgres") {
		urlCredMatcher = regexp.MustCompile(postgresDbURLPattern)
		if urlCredMatcher.MatchString(dbUrl) {
			redactUrl = urlCredMatcher.ReplaceAllString(dbUrl, `$1://$2:*REDACTED*@$4$5/$6`)
		}
	} else {
		urlCredMatcher = regexp.MustCompile(mysqlDbURLPattern)
		if urlCredMatcher.MatchString(dbUrl) {
			redactUrl = urlCredMatcher.ReplaceAllString(dbUrl, `$1:*REDACTED*@tcp($3$4)/$5`)
		}
	}
	return redactUrl
}
