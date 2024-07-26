package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	. "acceptance/helpers"

	"github.com/cloudfoundry/cf-test-helpers/v2/cf"
	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

const (
	PolicyPath           = "/v1/apps/%s/policy"
	AggregatedMetricPath = "/v1/apps/%s/aggregated_metric_histories/%s"
	HistoryPath          = "/v1/apps/%s/scaling_histories"
)

type AppInstanceMetric struct {
	AppId         string `json:"app_id"`
	InstanceIndex uint32 `json:"instance_index"`
	CollectedAt   int64  `json:"collected_at"`
	Name          string `json:"name"`
	Unit          string `json:"unit"`
	Value         string `json:"value"`
	Timestamp     int64  `json:"timestamp"`
}

type AppMetric struct {
	AppId      string `json:"app_id"`
	MetricType string `json:"name"`
	Value      string `json:"value"`
	Unit       string `json:"unit"`
	Timestamp  int64  `json:"timestamp"`
}

type AppScalingHistory struct {
	AppId        string `json:"app_id"`
	Timestamp    int64  `json:"timestamp"`
	ScalingType  int    `json:"scaling_type"`
	Status       int    `json:"status"`
	OldInstances int    `json:"old_instances"`
	NewInstances int    `json:"new_instances"`
	Reason       string `json:"reason"`
	Message      string `json:"message"`
	Error        string `json:"error"`
}

type MetricsResults struct {
	TotalResults uint32               `json:"total_results"`
	TotalPages   uint16               `json:"total_pages"`
	Page         uint16               `json:"page"`
	Metrics      []*AppInstanceMetric `json:"resources"`
}

type AggregatedMetricsResults struct {
	TotalResults uint32       `json:"total_results"`
	TotalPages   uint16       `json:"total_pages"`
	Page         uint16       `json:"page"`
	Metrics      []*AppMetric `json:"resources"`
}

type HistoryResults struct {
	TotalResults uint32               `json:"total_results"`
	TotalPages   uint16               `json:"total_pages"`
	Page         uint16               `json:"page"`
	Histories    []*AppScalingHistory `json:"resources"`
}

var (
	oauthToken string
)

var _ = Describe("AutoScaler Public API", func() {

	BeforeEach(func() {
		oauthToken = OauthToken(cfg)
		Expect(appGUID).NotTo(BeEmpty())
	})

	Context("when no policy defined", func() {

		BeforeEach(func() {
			_, status := deletePolicy()
			Expect(status).To(Or(Equal(200), Equal(404)))
		})

		It("should fail with 404 when retrieve policy", func() {
			_, status := getPolicy()
			Expect(status).To(Equal(404))
		})

		It("should succeed to create a valid policy", func() {
			policy := GenerateDynamicScaleOutPolicy(1, 2, "memoryused", 30)
			newPolicy, status := createPolicy(policy)
			Expect(status).To(Or(Equal(200), Equal(201)))
			Expect(string(newPolicy)).Should(MatchJSON(policy))
		})

		It("should succeed to create a valid policy but remove any extra fields", func() {
			policyWithExtraFields, validPolicy := GenerateDynamicScaleOutPolicyWithExtraFields(1, 2, "memoryused", 30)
			newPolicy, status := createPolicy(policyWithExtraFields)
			Expect(status).To(Or(Equal(200), Equal(201)))
			Expect(string(newPolicy)).ShouldNot(MatchJSON(policyWithExtraFields))
			Expect(string(newPolicy)).Should(MatchJSON(validPolicy))
		})

		It("should fail to create an invalid policy", func() {
			response, status := createPolicy(GenerateDynamicScaleOutPolicy(0, 2, "memoryused", 30))
			Expect(status).To(Equal(400))
			Expect(string(response)).Should(ContainSubstring(`[{"context":"(root).instance_min_count","description":"Must be greater than or equal to 1"}]`))
		})

	})

	Context("When policy is defined", func() {
		memThreshold := int64(10)
		var policy string

		BeforeEach(func() {
			policy = GenerateDynamicScaleOutPolicy(1, 2, "memoryused", memThreshold)
			_, status := createPolicy(policy)
			Expect(status).To(Or(Equal(200), Equal(201)))
		})

		It("should succeed to delete a policy", func() {
			_, status := deletePolicy()
			Expect(status).To(Equal(200))
		})

		It("should succeed to get a policy", func() {
			gotPolicy, status := getPolicy()
			Expect(status).To(Equal(200))
			Expect(string(gotPolicy)).Should(MatchJSON(policy))
		})

		It("should succeed to update a valid policy", func() {
			newPolicy, status := createPolicy(GenerateDynamicScaleOutPolicy(1, 2, "memoryused", memThreshold))
			Expect(status).To(Equal(200))
			Expect(string(newPolicy)).Should(MatchJSON(policy))
		})

		It("should succeed to update a valid policy but remove any extra fields", func() {
			policyWithExtraFields, validPolicy := GenerateDynamicScaleOutPolicyWithExtraFields(1, 2, "memoryused", memThreshold)
			newPolicy, status := createPolicy(policyWithExtraFields)
			Expect(status).To(Or(Equal(200), Equal(201)))
			Expect(string(newPolicy)).ShouldNot(MatchJSON(policyWithExtraFields))
			Expect(string(newPolicy)).Should(MatchJSON(validPolicy))
		})

		It("should fail to update an invalid policy", func() {
			By("return 400 when the new policy is invalid")
			_, status := createPolicy(GenerateDynamicScaleOutPolicy(0, 2, "memoryused", 30))
			Expect(status).To(Equal(400))

			By("the original policy is not changed")
			existing, status := getPolicy()
			Expect(status).To(Equal(200))
			Expect(string(existing)).Should(MatchJSON(policy))

		})

		Context("for an unrelated user", func() {
			BeforeEach(func() {
				workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
					// Make "other user" a space auditor in the space along with a space developer in the other space
					cmd := cf.Cf("set-space-role", otherSetup.RegularUserContext().Username, setup.RegularUserContext().Org, setup.RegularUserContext().Space, "SpaceAuditor")
					Expect(cmd.Wait(cfg.DefaultTimeoutDuration())).To(Exit(0))
				})
				workflowhelpers.AsUser(otherSetup.RegularUserContext(), cfg.DefaultTimeoutDuration(), func() {
					oauthToken = OauthToken(cfg)
				})
			})

			It("should not be possible to read the policy", func() {
				_, status := getPolicy()
				Expect(status).To(Equal(401))
			})
		})

		Context("When scale out is triggered", func() {
			BeforeEach(func() {
				totalTime := time.Duration(cfg.AggregateInterval*2)*time.Second + 3*time.Minute
				WaitForNInstancesRunning(appGUID, 2, totalTime)
			})

			It("should successfully scale out", func() {
				By("check aggregated metrics")
				Expect(len(getAggregatedMetrics().Metrics)).Should(BeNumerically(">=", 1))

				By("check history has scale event")
				for _, entry := range getHistory().Histories {
					Expect(entry.AppId).To(Equal(appGUID))
					Expect(entry.ScalingType).Should(BeNumerically("==", 0))
					Expect(entry.Status).Should(BeNumerically("==", 0))
					Expect(entry.Reason).To(Equal(fmt.Sprintf("+1 instance(s) because memoryused >= %dMB for %d seconds", memThreshold, TestBreachDurationSeconds)))
				}
			})
		})
	})
})

func getHistory() *HistoryResults {
	historyURL, err := url.Parse(cfg.ASApiEndpoint)
	Expect(err).ShouldNot(HaveOccurred())

	historyURL.Path = fmt.Sprintf(HistoryPath, appGUID)

	raw, status := get(historyURL.String())
	Expect(status).To(Equal(200))

	var histories *HistoryResults
	err = json.Unmarshal(raw, &histories)
	Expect(err).ShouldNot(HaveOccurred())
	return histories
}

func createPolicy(policy string) ([]byte, int) {
	policyURL, err := url.Parse(cfg.ASApiEndpoint)
	Expect(err).ShouldNot(HaveOccurred())
	policyURL.Path = fmt.Sprintf(PolicyPath, appGUID)

	return put(policyURL.String(), policy)
}

func put(url string, body string) ([]byte, int) {
	By(fmt.Sprintf("PUT '%s'", url))
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer([]byte(body)))
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	Expect(err).ShouldNot(HaveOccurred())

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	Expect(err).ShouldNot(HaveOccurred())
	return raw, resp.StatusCode
}

func deletePolicy() ([]byte, int) {
	policyURL, err := url.Parse(cfg.ASApiEndpoint)
	Expect(err).ShouldNot(HaveOccurred())
	policyURL.Path = fmt.Sprintf(PolicyPath, appGUID)
	return deleteReq(policyURL.String())
}

func deleteReq(url string) ([]byte, int) {
	//delete policy here to make sure the condtion "no policy defined"
	req, err := http.NewRequest("DELETE", url, nil)
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	Expect(err).ShouldNot(HaveOccurred())
	defer func() { _ = resp.Body.Close() }()
	response, err := io.ReadAll(resp.Body)
	Expect(err).ShouldNot(HaveOccurred())
	return response, resp.StatusCode
}

func getAggregatedMetrics() *AggregatedMetricsResults {
	var metrics *AggregatedMetricsResults
	aggregatedMetricURL, err := url.Parse(cfg.ASApiEndpoint)
	aggregatedMetricURL.Path = fmt.Sprintf(AggregatedMetricPath, appGUID, "memoryused")

	raw, status := get(aggregatedMetricURL.String())
	Expect(status).To(Equal(200))

	err = json.Unmarshal(raw, &metrics)
	Expect(err).ShouldNot(HaveOccurred())
	return metrics
}

func getPolicy() ([]byte, int) {
	policyURL, err := url.Parse(cfg.ASApiEndpoint)
	Expect(err).ShouldNot(HaveOccurred())
	policyURL.Path = fmt.Sprintf(PolicyPath, appGUID)

	return get(policyURL.String())
}

func get(url string) ([]byte, int) {
	req, err := http.NewRequest("GET", url, nil)
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	Expect(err).ShouldNot(HaveOccurred())

	defer func() { _ = resp.Body.Close() }()

	policy, err := io.ReadAll(resp.Body)
	Expect(err).ShouldNot(HaveOccurred())
	return policy, resp.StatusCode
}
