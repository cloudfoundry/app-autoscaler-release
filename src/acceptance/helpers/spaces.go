package helpers

import (
	"acceptance/config"
	"encoding/json"
	"fmt"
	"github.com/KevinJCross/cf-test-helpers/v2/cf"
	"github.com/onsi/ginkgo/v2"
	"net/url"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type SpaceResources struct {
	Resources []Space `json:"resources"`
}
type Space struct {
	Guid string `json:"guid"`
	Name string `json:"name"`
}

func GetTestSpaces(orgGuid string, cfg *config.Config) []string {
	rawSpaces := GetRawSpaces(orgGuid, cfg.DefaultTimeoutDuration())

	var spaceNames []string
	for _, space := range rawSpaces {
		if strings.HasPrefix(space.Name, cfg.NamePrefix) {
			spaceNames = append(spaceNames, space.Name)
		}
	}
	ginkgo.GinkgoWriter.Printf("\nGot orgs: %s\n", spaceNames)
	return spaceNames
}

func GetRawSpaces(orgGuid string, timeout time.Duration) []Space {
	params := url.Values{"organization_guids": []string{orgGuid}}
	rawSpaces := cf.CfSilent("curl", fmt.Sprintf("/v3/spaces?%s", params.Encode())).Wait(timeout)
	Expect(rawSpaces).To(Exit(0), "unable to get spaces", timeout)
	spaces := SpaceResources{}
	err := json.Unmarshal(rawSpaces.Out.Contents(), &spaces)
	Expect(err).ShouldNot(HaveOccurred())
	return spaces.Resources
}

func DeleteSpaces(orgName string, spaces []string, timeout time.Duration) {
	if len(spaces) == 0 {
		return
	}

	fmt.Println(fmt.Sprintf("\nDeleting spaces: %s ", strings.Join(spaces, ", ")))

	for _, spaceName := range spaces {
		if timeout.Seconds() == 0 {
			deleteSpace := cf.Cf("delete-space", "-f", "-o", orgName, spaceName).Wait(timeout)
			Expect(deleteSpace).To(Exit(0), fmt.Sprintf("failed deleting space: %s in org: %s: %s", spaceName, orgName, string(deleteSpace.Err.Contents())))
		} else {
			cf.Cf("delete-space", "-f", "-o", orgName, spaceName)
		}
	}
}
