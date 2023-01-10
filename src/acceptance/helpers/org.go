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

func GetOrgQuotaNameFrom(orgGuid string, timeout time.Duration) string {
	var quota cfResourceObject
	rawQuota := cf.Cf("curl", "/v3/organization_quotas?organization_guids="+orgGuid).Wait(timeout)
	Expect(rawQuota).To(Exit(0), "unable to get services")
	err := json.Unmarshal(rawQuota.Out.Contents(), &quota)
	Expect(err).NotTo(HaveOccurred())
	return quota.Resources[0].Name
}

func GetTestOrgs(cfg *config.Config) []string {
	rawOrgs := getRawOrgs(cfg.DefaultTimeoutDuration())

	var orgNames []string
	for _, org := range rawOrgs {
		if strings.HasPrefix(org.Name, cfg.NamePrefix) {
			orgNames = append(orgNames, org.Name)
		}
	}
	ginkgo.GinkgoWriter.Printf("\nGot orgs: %s\n", orgNames)
	return orgNames
}

func DeleteOrgs(orgs []string, timeout time.Duration) {
	if len(orgs) == 0 {
		return
	}

	fmt.Println(fmt.Sprintf("Deleting orgs: %s ", strings.Join(orgs, ", ")))

	for _, org := range orgs {
		deleteOrg := cf.Cf("delete-org", org, "-f").Wait(timeout)
		Expect(deleteOrg).To(Exit(0), fmt.Sprintf("unable to delete org %s", org))
	}
}

func GetOrgGuid(cfg *config.Config, org string) string {
	orgGuidByte := cf.Cf("org", org, "--guid").Wait(cfg.DefaultTimeoutDuration())
	return strings.TrimSuffix(string(orgGuidByte.Out.Contents()), "\n")
}

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
