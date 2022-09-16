package helpers

import (
	"acceptance/config"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/onsi/ginkgo/v2"

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

func GetTestOrgs(cfg *config.Config) []string {
	rawOrgs := cf.CfSilent("curl", "/v3/organizations").Wait(cfg.DefaultTimeoutDuration())
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
	ginkgo.GinkgoWriter.Printf("\nGot orgs: %s\n", orgNames)
	return orgNames
}

type SpaceResources struct {
	Resources []Space `json:"resources"`
}
type Space struct {
	Guid string `json:"guid"`
	Name string `json:"name"`
}

func GetOrgSpaceNamesAndGuids(cfg *config.Config, org string) (orgName string, orgGuid string, spaceName string, spaceGuid string) {
	orgGuid = GetOrgGuid(cfg, org)
	spaces := GetSpaces(cfg, orgGuid)
	if len(spaces.Resources) == 0 {
		return org, orgGuid, "", ""
	}
	spaceName = spaces.Resources[0].Name
	spaceGuid = spaces.Resources[0].Guid

	ginkgo.GinkgoWriter.Printf("\nUsing Org: %s - %s\n", org, orgGuid)
	ginkgo.GinkgoWriter.Printf("\nUsing Space: %s - %s\n", spaceName, spaceGuid)
	return org, orgGuid, spaceName, spaceGuid
}

func GetSpaces(cfg *config.Config, orgGuid string) struct {
	Resources []Space `json:"resources"`
} {
	params := url.Values{"organization_guids": []string{orgGuid}}
	rawSpaces := cf.CfSilent("curl", fmt.Sprintf("/v3/spaces?%s", params.Encode())).Wait(cfg.DefaultTimeoutDuration())
	Expect(rawSpaces).To(Exit(0), "unable to get spaces")
	spaces := SpaceResources{}
	err := json.Unmarshal(rawSpaces.Out.Contents(), &spaces)
	Expect(err).ShouldNot(HaveOccurred())
	return spaces
}

func GetOrgGuid(cfg *config.Config, org string) string {
	orgGuidByte := cf.Cf("org", org, "--guid").Wait(cfg.DefaultTimeoutDuration())
	return strings.TrimSuffix(string(orgGuidByte.Out.Contents()), "\n")
}

func DeleteOrg(cfg *config.Config, org string) {
	deleteOrg := cf.Cf("delete-org", org, "-f").Wait(cfg.DefaultTimeoutDuration())
	Expect(deleteOrg).To(Exit(0), fmt.Sprintf("unable to delete org %s", org))
}
