package helpers

import (
	"encoding/json"
	"time"

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
