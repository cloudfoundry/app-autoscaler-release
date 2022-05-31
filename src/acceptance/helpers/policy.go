package helpers

import (
	"code.cloudfoundry.org/app-autoscaler/src/acceptance/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/onsi/gomega"
)

func GetPolicy(cfg *config.Config, appGUID string) ScalingPolicy {
	policyURL := fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(PolicyPath, "{appId}", appGUID, -1))
	oauthToken := OauthToken(cfg)
	client := GetHTTPClient(cfg)

	req, err := http.NewRequest("GET", policyURL, nil)
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	Expect(err).ShouldNot(HaveOccurred())

	defer func() { _ = resp.Body.Close() }()

	raw, err := ioutil.ReadAll(resp.Body)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))

	var responsedPolicy ScalingPolicy
	err = json.Unmarshal(raw, &responsedPolicy)
	Expect(err).ShouldNot(HaveOccurred())
	return responsedPolicy
}
