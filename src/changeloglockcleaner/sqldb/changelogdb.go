package sqldb

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"

	_ "github.com/lib/pq"
)

const PostgresDriverName = "postgres"
const postgresDbURLPattern = `^(postgres|postgresql):\/\/(.+):(.+)@([\da-zA-Z\.-]+)(:[\d]{4,5})?\/(.+)`

type ChangelogSQLDB struct {
	sqldb *sql.DB
}

func NewChangelogSQLDB(dbUrl string) (*ChangelogSQLDB, error) {
	log.SetOutput(os.Stdout)
	sqldb, err := sql.Open(PostgresDriverName, dbUrl)
	if err != nil {
		return nil, err
	}

	err = sqldb.Ping()
	if err != nil {
		sqldb.Close()
		urlCredMatcher := regexp.MustCompile(postgresDbURLPattern)
		if urlCredMatcher.MatchString(dbUrl) {
			dbUrl = urlCredMatcher.ReplaceAllString(dbUrl, `$1://$2:*REDACTED*@$4$5/$6`)
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
}
