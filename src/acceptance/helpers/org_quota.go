package helpers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/KevinJCross/cf-test-helpers/v2/cf"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type OrgQuota struct {
	Name             string
	TotalMemory      string
	InstanceMemory   string
	Routes           string
	ServiceInstances string
	AppInstances     string
	RoutePorts       string
}

func UpdateOrgQuota(orgQuota OrgQuota, timeout time.Duration) {
	args := []string{"update-org-quota", orgQuota.Name}
	args = append(args, "-i", orgQuota.AppInstances)
	args = append(args, "-r", orgQuota.Routes)
	args = append(args, "-s", orgQuota.ServiceInstances)
	args = append(args, "-m", orgQuota.TotalMemory)
	args = append(args, "--reserved-route-ports", orgQuota.RoutePorts)
	updateOrgQuota := cf.Cf(args...).Wait(timeout)
	Expect(updateOrgQuota).To(Exit(0), "unable update org quota: "+string(updateOrgQuota.Out.Contents()[:]))
}

func GetOrgQuota(orgGuid string, timeout time.Duration) (orgQuota OrgQuota) {
	rawQuota := getRawOrgQuota(orgGuid, timeout).Resources[0]
	orgQuota = OrgQuota{
		Name:           rawQuota.Name,
		TotalMemory:    fmt.Sprint("%sMB", rawQuota.Apps.TotalMemoryInMb),
		InstanceMemory: fmt.Sprintf("%sMB", rawQuota.Apps.PerProcessMemoryInMb),
	}

	if rawQuota.Routes.TotalRoutes != 0 {
		orgQuota.Routes = string(rawQuota.Routes.TotalRoutes)
	}

	if rawQuota.Services.TotalServiceInstances != 0 {
		orgQuota.ServiceInstances = string(rawQuota.Services.TotalServiceInstances)
	}

	if rawQuota.Apps.TotalInstances != 0 {
		orgQuota.AppInstances = string(rawQuota.Apps.TotalInstances)
	}

	if rawQuota.Routes.TotalRoutes != 0 {
		orgQuota.Routes = string(rawQuota.Routes.TotalRoutes)
	}

	return orgQuota
}

func getRawOrgQuota(orgGuid string, timeout time.Duration) cfResourceObject {
	var quota cfResourceObject
	rawQuota := cf.Cf("curl", "/v3/organization_quotas?organization_guids="+orgGuid).Wait(timeout)
	Expect(rawQuota).To(Exit(0), "unable to get services")
	err := json.Unmarshal(rawQuota.Out.Contents(), &quota)
	Expect(err).NotTo(HaveOccurred())
	return quota
}
