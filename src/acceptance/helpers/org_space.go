package helpers

import (
	"acceptance/config"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/KevinJCross/cf-test-helpers/v2/cf"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type cfOrgs struct {
	Resources []struct {
		Name      string `json:"name"`
		GUID      string `json:"guid"`
		CreatedAt string `json:"created_at"`
	} `json:"resources"`
}

type cfSpaces struct {
	Resources []struct {
		Entity struct {
			Name string `json:"name"`
		}
		Metadata struct {
			GUID      string `json:"guid"`
			CreatedAt string `json:"created_at"`
		}
	} `json:"resources"`
}

func GetTestOrgs(cfg *config.Config) []string {
	rawOrgs := cf.Cf("curl", "/v3/organizations").Wait(cfg.DefaultTimeoutDuration())
	Expect(rawOrgs).To(Exit(0), "unable to get orgs")

	var orgs cfOrgs
	err := json.Unmarshal(rawOrgs.Out.Contents(), &orgs)
	Expect(err).ShouldNot(HaveOccurred())

	var orgNames []string
	for _, org := range orgs.Resources {
		if strings.HasPrefix(org.Name, cfg.NamePrefix) {
			orgNames = append(orgNames, org.Name)
		}
	}

	return orgNames
}

func GetOrgSpaceNamesAndGuids(cfg *config.Config, org string) (string, string, string, string) {
	orgGuidByte := cf.Cf("org", org, "--guid").Wait(cfg.DefaultTimeoutDuration())
	orgGuid := strings.TrimSuffix(string(orgGuidByte.Out.Contents()), "\n")

	rawSpaces := cf.Cf("curl", fmt.Sprintf("/v2/organizations/%s/spaces", orgGuid)).Wait(cfg.DefaultTimeoutDuration())
	Expect(rawSpaces).To(Exit(0), "unable to get spaces")
	var spaces cfSpaces
	err := json.Unmarshal(rawSpaces.Out.Contents(), &spaces)
	Expect(err).ShouldNot(HaveOccurred())
	if len(spaces.Resources) == 0 {
		return org, orgGuid, "", ""
	}

	return org, orgGuid, spaces.Resources[0].Entity.Name, spaces.Resources[0].Metadata.GUID
}

func DeleteOrg(cfg *config.Config, org string) {
	deleteOrg := cf.Cf("delete-org", org, "-f").Wait(cfg.DefaultTimeoutDuration())
	Expect(deleteOrg).To(Exit(0), fmt.Sprintf("unable to delete org %s", org))
}
