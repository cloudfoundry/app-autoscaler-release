package helpers

import (
	"acceptance/config"
	"encoding/json"
	"fmt"
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
func FindExistingOrgAndSpace(cfg *config.Config) (orgName string, spaceName string) {
	organizations := GetTestOrgs(cfg)
	Expect(len(organizations)).To(Equal(1))
	orgName = organizations[0]
	_, _, spaceName, _ = GetOrgSpaceNamesAndGuids(cfg, orgName)

	return orgName, spaceName
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

func GetOrgSpaceNamesAndGuids(cfg *config.Config, org string) (orgName string, orgGuid string, spaceName string, spaceGuid string) {
	orgGuid = GetOrgGuid(cfg, org)
	spaces := GetRawSpaces(orgGuid, cfg.DefaultTimeoutDuration())
	if len(spaces) == 0 {
		return org, orgGuid, "", ""
	}
	spaceName = spaces[0].Name
	spaceGuid = spaces[0].Guid

	ginkgo.GinkgoWriter.Printf("\nUsing Org: %s - %s\n", org, orgGuid)
	ginkgo.GinkgoWriter.Printf("\nUsing Space: %s - %s\n", spaceName, spaceGuid)
	return org, orgGuid, spaceName, spaceGuid
}

func GetOrgGuid(cfg *config.Config, org string) string {
	orgGuidByte := cf.Cf("org", org, "--guid").Wait(cfg.DefaultTimeoutDuration())
	return strings.TrimSuffix(string(orgGuidByte.Out.Contents()), "\n")
}

func DeleteOrgWithTimeout(org string, timeout time.Duration) {
	deleteOrg := cf.Cf("delete-org", org, "-f").Wait(timeout)
	Expect(deleteOrg).To(Exit(0), fmt.Sprintf("unable to delete org %s", org))
}
