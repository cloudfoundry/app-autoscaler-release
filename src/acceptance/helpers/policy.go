package helpers

import (
	"acceptance/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	. "github.com/onsi/gomega"
)

func GetPolicy(cfg *config.Config, appGUID string) ScalingPolicy {
	policyURL, err := url.Parse(cfg.ASApiEndpoint)
	Expect(err).ShouldNot(HaveOccurred())
	policyURL.Path = fmt.Sprintf("/v1/apps/%s/policy", appGUID)

	oauthToken := OauthToken(cfg)
	client := GetHTTPClient(cfg)

	req, err := http.NewRequest("GET", policyURL.String(), nil)
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	Expect(err).ShouldNot(HaveOccurred())

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))

	var responsedPolicy ScalingPolicy
	err = json.Unmarshal(raw, &responsedPolicy)
	Expect(err).ShouldNot(HaveOccurred())
	return responsedPolicy
}
