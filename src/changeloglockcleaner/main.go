package main

import (
	"flag"
	"fmt"
	"os"

	"changeloglockcleaner/sqldb"
)

func main() {
	var dbUrl string
	var lockExpirationDurationInSeconds int

	flag.StringVar(&dbUrl, "dburl", "", "dburl")
	flag.IntVar(&lockExpirationDurationInSeconds, "expiration_second", 0, "lock expiration duration in seconds")
	flag.Parse()
	if dbUrl == "" {
		showErrorAndUsage("dburl")
		os.Exit(1)
	}
	if lockExpirationDurationInSeconds <= 0 {
		showErrorAndUsage("expiration_second")
		os.Exit(1)
	}
	db, err := sqldb.NewChangelogSQLDB(dbUrl)
	if err != nil {
		fmt.Fprintf(os.Stdout, "failed to connect to database:\n %s\n", err)
		os.Exit(1)
	}
	defer db.Close()
	err = db.DeleteExpiredLock(lockExpirationDurationInSeconds)
	if err != nil {
		fmt.Fprintf(os.Stdout, "failed to delete expired lock:\n %s\n", err)
		os.Exit(1)
	}
}
func showErrorAndUsage(missingOpt string) {
	fmt.Fprintf(os.Stdout, `Incorrect Usage: '-%s' is missing
		
	Usage:
	changeloglockcleaner -dburl DBRUL -expiration_second DURATION
	
	Options:
	-dburl The URL of database
	-expiration_second Lock expiration duration in seconds`, missingOpt)
}
