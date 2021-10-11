package helpers

import (
	"acceptance/config"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func GetApps(cfg *config.Config, orgGuid, spaceGuid string, prefix string) []string {
	var apps cfResourceObject
	rawApps := cf.Cf("curl", "/v3/apps?space_guids="+spaceGuid+"&organization_guids="+orgGuid).Wait(cfg.DefaultTimeoutDuration())
	Expect(rawApps).To(Exit(0), "unable to get apps")
	err := json.Unmarshal(rawApps.Out.Contents(), &apps)
	Expect(err).ShouldNot(HaveOccurred())

	var names []string
	for _, app := range apps.Resources {
		if strings.HasPrefix(app.Name, prefix) {
			names = append(names, app.Name)
		}
	}

	return names
}

func DeleteApps(cfg *config.Config, apps []string, threshold int) {
	for _, app := range apps {
		deleteApp := cf.Cf("delete", app, "-f").Wait(cfg.DefaultTimeoutDuration())
		Expect(deleteApp).To(Exit(0), fmt.Sprintf("unable to delete app %s", app))
	}
}
