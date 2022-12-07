package helpers

import (
	"acceptance/config"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"

	"github.com/KevinJCross/cf-test-helpers/v2/cf"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func getRawOrgsByPage(page int, timeout time.Duration) cfResourceObject {
	var response cfResourceObject
	rawResponse := cf.Cf("curl", "/v3/organizations?&page="+strconv.Itoa(page)).Wait(timeout)
	Expect(rawResponse).To(Exit(0), "unable to get orgs")
	err := json.Unmarshal(rawResponse.Out.Contents(), &response)
	Expect(err).ShouldNot(HaveOccurred())
	return response
}

func getRawOrgs(timeout time.Duration) []cfResource {
	var rawOrgs []cfResource
	totalPages := 1

	for page := 1; page <= totalPages; page++ {
		var response = getRawOrgsByPage(page, timeout)
		totalPages = response.Pagination.TotalPages
		rawOrgs = append(rawOrgs, response.Resources...)
	}

	return rawOrgs
}

func GetTestOrgs(cfg *config.Config) []string {
	rawOrgs := getRawOrgs(cfg.DefaultTimeoutDuration())

	var orgNames []string
	for _, org := range rawOrgs {
		if strings.HasPrefix(org.Name, cfg.NamePrefix) || org.Name == cfg.ExistingOrganization {
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

func DeleteOrgWithTimeout(org string, timeout time.Duration) {
	deleteOrg := cf.Cf("delete-org", org, "-f").Wait(timeout)
	Expect(deleteOrg).To(Exit(0), fmt.Sprintf("unable to delete org %s", org))
}
