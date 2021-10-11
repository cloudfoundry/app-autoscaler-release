package helpers

import (
	"acceptance/config"
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type cfResourceObject struct {
	Resources []struct {
		GUID      string `json:"guid"`
		CreatedAt string `json:"created_at"`
		Name      string `json:"name"`
		Username  string `json:"username"`
	} `json:"resources"`
}

func GetServices(cfg *config.Config, orgGuid, spaceGuid string, prefix string) []string {
	var services cfResourceObject
	rawServices := cf.Cf("curl", "/v3/service_instances?space_guids="+spaceGuid+"&organization_guids="+orgGuid).Wait(cfg.DefaultTimeoutDuration())
	Expect(rawServices).To(Exit(0), "unable to get services")
	err := json.Unmarshal(rawServices.Out.Contents(), &services)
	Expect(err).ShouldNot(HaveOccurred())

	return filterByPrefix(prefix, getNames(services))
}

func DeleteServices(cfg *config.Config, services []string) {
	for _, service := range services {
		deleteService := cf.Cf("delete-service", service, "-f").Wait(cfg.DefaultTimeoutDuration())
		if deleteService.ExitCode() != 0 {
			fmt.Printf("unable to delete the service %s, attempting to purge...\n", service)
			purgeService := cf.Cf("purge-service-instance", service, "-f").Wait(cfg.DefaultTimeoutDuration())
			Expect(purgeService).To(Exit(0), fmt.Sprintf("unable to delete service %s", service))
		}
	}
}
