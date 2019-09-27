package sqldb_test

import (
	"changeloglockcleaner/sqldb"

	"github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"
)

var _ = Describe("ChangelogSQLDB", func() {
	var (
		cdb              *sqldb.ChangelogSQLDB
		timeoutInSecound int = 300
		err              error
		dbUrl            string
	)
	Describe("NewChangelogSQLDB", func() {
		JustBeforeEach(func() {
			cdb, err = sqldb.NewChangelogSQLDB(dbUrl)
		})
		BeforeEach(func() {
			dbUrl = os.Getenv("DBURL")
		})
		AfterEach(func() {
			if cdb != nil {
				err = cdb.Close()
				Expect(err).NotTo(HaveOccurred())
			}
		})

		Context("when db url is not correct", func() {
			BeforeEach(func() {
				dbUrl = "postgres://not-exist-user:not-exist-password@localhost/autoscaler?sslmode=disable"
			})
			It("should error", func() {
				Expect(err).To(BeAssignableToTypeOf(&pq.Error{}))
			})

		})

		Context("when db url is correct", func() {
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(cdb).NotTo(BeNil())
			})
		})
	})

	Describe("DeleteExpiredLock", func() {
		BeforeEach(func() {
			cdb, err = sqldb.NewChangelogSQLDB(os.Getenv("DBURL"))
			Expect(err).NotTo(HaveOccurred())
			cleanChanglogLockTable()
		})

		AfterEach(func() {
			err = cdb.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			err = cdb.DeleteExpiredLock(timeoutInSecound)
		})

		Context("when the lock is not expired", func() {
			BeforeEach(func() {
				_, err := insertLock(1, true, (0 - timeoutInSecound + 60), "locker")
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not error", func() {
				Expect(checkChanglogLockExistenceById(1)).To(BeTrue())
			})
		})
		Context("when the lock is expired", func() {
			BeforeEach(func() {
				_, err := insertLock(1, true, (0 - timeoutInSecound - 60), "locker")
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not error", func() {
				Expect(checkChanglogLockExistenceById(1)).To(BeFalse())
			})
		})

	})

})
