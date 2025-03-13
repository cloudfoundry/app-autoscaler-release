package testhelpers

import (
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
)

func GetDbVcapServices(creds map[string]string, serviceName string, dbType string) (string, error) {
	credentials, err := json.Marshal(creds)
	if err != nil {
		return "", err
	}

	return `{
		"user-provided": [ { "name": "config", "credentials": { "metricsforwarder": { } }}],
		"autoscaler": [ {
			"name": "some-service",
			"credentials": ` + string(credentials) + `,
			"syslog_drain_url": "",
			"tags": ["` + serviceName + `", "` + dbType + `"]
			}
		]}`, nil // #nosec G101
}

func GetVcapServices(userProvidedServiceName string, configJson string) string {
	GinkgoHelper()
	dbURL := os.Getenv("DBURL")

	catalogBytes, err := os.ReadFile("../api/exampleconfig/catalog-example.json")
	Expect(err).NotTo(HaveOccurred())

	return `{
		"user-provided": [
		    { "name": "` + userProvidedServiceName + `", "tags": [ "` + userProvidedServiceName + `" ],  "credentials": { "` + userProvidedServiceName + `": ` + configJson + ` }},
			{ "name": "broker-catalog", "tags": ["broker-catalog"], "credentials": { "broker-catalog": ` + string(catalogBytes) + ` }}
		],
		"autoscaler": [ {
			"name": "some-service",
			"credentials": {
				"uri": "` + dbURL + `"
				},
			"syslog_drain_url": "",
			"tags": [ "policy_db","binding_db", "postgres" ]

	   }]}`
}
