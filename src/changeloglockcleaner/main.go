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

	flag.StringVar(&dbUrl, "d", "", "dburl")
	flag.IntVar(&lockExpirationDurationInSeconds, "e", 0, "lock expiration duration in seconds")
	flag.Parse()
	if dbUrl == "" {
		showErrorAndUsage("d")
		os.Exit(1)
	}
	if lockExpirationDurationInSeconds <= 0 {
		showErrorAndUsage("e")
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
	changeloglockcleaner -d DBURL -e EXPIRATION_SECONDS
	
	Options:
	-d The URL of database
	-e Lock expiration duration in seconds`, missingOpt)
}
