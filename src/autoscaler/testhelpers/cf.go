package testhelpers

import (
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
)

func GetStoredProcedureDbVcapServices(creds map[string]string, serviceName string, dbType string) (string, error) {
	credentials, err := json.Marshal(creds)
	if err != nil {
		return "", err
	}

	return getDbVcapServices(string(credentials), serviceName, dbType, "stored_procedure")
}

func GetDbVcapServices(creds map[string]string, serviceName string, dbType string) (string, error) {
	credentials, err := json.Marshal(creds)
	if err != nil {
		return "", err
	}

	return getDbVcapServices(string(credentials), serviceName, dbType, "default")
}

func getDbVcapServices(creds, serviceName, dbType, credHelperImpl string) (string, error) {
	vcapServices := map[string]interface{}{
		"user-provided": []map[string]interface{}{
			{
				"name": "config",
				"credentials": map[string]interface{}{
					"metricsforwarder": map[string]interface{}{
						"cred_helper_impl": credHelperImpl,
					},
				},
			},
		},
		"autoscaler": []map[string]interface{}{
			{
				"name": "some-service",
				"credentials": json.RawMessage(creds),
				"syslog_drain_url": "",
				"tags": []string{serviceName, dbType},
			},
		},
	}

	vcapServicesJson, err := json.Marshal(vcapServices)
	if err != nil {
		return "", err
	}

	return string(vcapServicesJson), nil
}

func GetVcapServices(userProvidedServiceName string, configJson string) string {
	GinkgoHelper()
	dbURL := os.Getenv("DBURL")

	catalogBytes, err := os.ReadFile("../api/exampleconfig/catalog-example.json")
	Expect(err).NotTo(HaveOccurred())

	vcapServices := map[string]interface{}{
		"user-provided": []map[string]interface{}{
			{
				"name":        userProvidedServiceName,
				"tags":        []string{userProvidedServiceName},
				"credentials": map[string]interface{}{userProvidedServiceName: json.RawMessage(configJson)},
			},
			{
				"name":        "broker-catalog",
				"tags":        []string{"broker-catalog"},
				"credentials": map[string]interface{}{"broker-catalog": json.RawMessage(catalogBytes)},
			},
		},
		"autoscaler": []map[string]interface{}{
			{
				"name": "some-service",
				"credentials": map[string]interface{}{
					"uri": dbURL,
				},
				"syslog_drain_url": "",
				"tags":             []string{"policy_db", "binding_db", "postgres"},
			},
		},
	}

	vcapServicesJson, err := json.Marshal(vcapServices)
	Expect(err).NotTo(HaveOccurred())

	return string(vcapServicesJson)
}
