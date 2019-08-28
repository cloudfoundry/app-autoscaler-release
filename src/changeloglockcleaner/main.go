package main

import (
	"flag"
	"fmt"
	"os"

	"changeloglockcleaner/sqldb"
)

func main() {
	var dbUrl string
	var lockExpiredDuration int

	flag.StringVar(&dbUrl, "dburl", "", "dburl")
	flag.IntVar(&lockExpiredDuration, "expired_second", 0, "lock expired duration second")
	flag.Parse()
	if dbUrl == "" {
		fmt.Fprintln(os.Stdout, "missing dburl\nUsage:use '-dburl' option to specify the dburl")
		os.Exit(1)
	}
	if lockExpiredDuration <= 0 {
		fmt.Fprintln(os.Stdout, "missing expired_second\nUsage:use '-expired_second' option to specify the expired_second")
		os.Exit(1)
	}
	db, err := sqldb.NewChangelogSQLDB(dbUrl)
	if err != nil {
		fmt.Fprintf(os.Stdout, "failed to connect to database:\n %s\n", err)
		os.Exit(1)
	}
	defer db.Close()
	err = db.DeleteExpiredLock(lockExpiredDuration)
	if err != nil {
		fmt.Fprintf(os.Stdout, "failed to delete expired lock:\n %s\n", err)
		os.Exit(1)
	}
}
