package db_test

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/db"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

var _ = Describe("Config", func() {

	var (
		conf        *db.DatabaseConfig
		err         error
		configBytes []byte
	)

	Describe("LoadConfig", func() {

		JustBeforeEach(func() {
			dec := yaml.NewDecoder(bytes.NewBuffer(configBytes))
			dec.KnownFields(true)
			err = dec.Decode(&conf)
		})

		// this test demonstrates that JSON configuration parsing is broken at the moment
		When("Configuration is provided as JSON", func() {
			It("Should be parseable", func() {
				var dbConfig db.DatabaseConfig
				configBytes := []byte(`
{
  "max_open_connections": 1
}
`)
				err := json.Unmarshal(configBytes, &dbConfig)

				Expect(err).NotTo(HaveOccurred())
				Expect(dbConfig.MaxOpenConnections).To(Equal(int32(1)))
			})
		})

		Context("when the config is valid", func() {
			BeforeEach(func() {
				configBytes = []byte(`
url: "postgres://postgres:postgres@localhost/autoscaler?sslmode=disable"
max_open_connections: 5
max_idle_connections: 5
connection_max_lifetime: 60s
connection_max_idletime: 60s
`)
			})

			It("should succeed", func() {
				Expect(err).NotTo(HaveOccurred())

				Expect(conf).To(Equal(&db.DatabaseConfig{
					URL:                   "postgres://postgres:postgres@localhost/autoscaler?sslmode=disable",
					MaxOpenConnections:    5,
					MaxIdleConnections:    5,
					ConnectionMaxLifetime: 60000000000,
					ConnectionMaxIdleTime: 60000000000,
				}))
			})
		})

		Context("when it gives a non integer max_open_connections", func() {
			BeforeEach(func() {
				configBytes = []byte(`max_open_connections: NOT-INTEGER-VALUE`)
			})

			It("should error", func() {
				Expect(err).To(BeAssignableToTypeOf(&yaml.TypeError{}))
				Expect(err).To(MatchError(MatchRegexp("cannot unmarshal .* into int")))
			})
		})

		Context("when it gives a non integer max_idle_connections", func() {
			BeforeEach(func() {
				configBytes = []byte(`max_idle_connections: NOT-INTEGER-VALUE`)
			})

			It("should error", func() {
				Expect(err).To(BeAssignableToTypeOf(&yaml.TypeError{}))
				Expect(err).To(MatchError(MatchRegexp("cannot unmarshal .* into int")))
			})
		})

		Context("when connection_max_lifetime is not a time duration", func() {
			BeforeEach(func() {
				configBytes = []byte(`connection_max_lifetime: 60k`)
			})

			It("should error", func() {
				Expect(err).To(BeAssignableToTypeOf(&yaml.TypeError{}))
				Expect(err).To(MatchError(MatchRegexp("cannot unmarshal .* into time.Duration")))
			})
		})

		Context("when connection_max_idletime is not a time duration", func() {
			BeforeEach(func() {
				configBytes = []byte(`connection_max_idletime: 60k`)
			})

			It("should error", func() {
				Expect(err).To(BeAssignableToTypeOf(&yaml.TypeError{}))
				Expect(err).To(MatchError(MatchRegexp("cannot unmarshal .* into time.Duration")))
			})
		})
	})
})
