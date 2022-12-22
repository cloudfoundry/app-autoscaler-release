package sqldb_test

import (
	. "changeloglockcleaner/sqldb"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"strings"
)

var _ = Describe("Helper", func() {
	var (
		dbUrl    string
		err      error
		database *Database
		dbHost   = os.Getenv("DB_HOST")
	)

	Describe("GetConnection", func() {

		JustBeforeEach(func() {
			database, err = GetConnection(dbUrl)
		})
		Context("when mysql query parameters are provided", func() {
			BeforeEach(func() {
				if strings.Contains(dbUrl, "postgres") {
					Skip("Not configured for mysql")
				}
				dbUrl = "root@tcp(" + dbHost + ":3306)/autoscaler?tls=preferred"
			})
			It("returns mysql database object", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(database).To(Equal(&Database{
					DriverName: "mysql",
					DSN:        "root@tcp(" + dbHost + ":3306)/autoscaler?parseTime=true&tls=preferred",
				}))
			})
		})

		Context("when mysql query parameters are not provided", func() {
			BeforeEach(func() {
				dbUrl = "root@tcp(localhost:3306)/autoscaler"
			})
			It("returns mysql database object", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(database).To(Equal(&Database{
					DriverName: "mysql",
					DSN:        "root@tcp(localhost:3306)/autoscaler?parseTime=true",
				}))
			})

		})

		Context("when need to verify mysql server, cert is not provided ", func() {
			BeforeEach(func() {
				if strings.Contains(dbUrl, "postgres") {
					Skip("Not configured for mysql")
				}
				dbUrl = "root@tcp(" + dbHost + ":3306)/autoscaler?tls=verify-ca"
			})
			It("should error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("sql ca file is not provided"))
			})
		})

		Context("when postgres dburl is provided", func() {
			BeforeEach(func() {
				dbUrl = "postgres://postgres:password@"+dbHost+":5432/autoscaler?sslmode=disable"
			})
			It("returns postgres database object", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(database).To(Equal(&Database{
					DriverName: "pgx",
					DSN:        "postgres://postgres:password@"+dbHost+":5432/autoscaler?sslmode=disable",
				}))
			})
		})
	})
})
