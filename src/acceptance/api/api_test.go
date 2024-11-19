package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	. "acceptance/helpers"

	"github.com/cloudfoundry/cf-test-helpers/v2/cf"
	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
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

	When("no scaling policy is set", func() {

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

		// custom metrics strategy - FIXME This test can be removed as it is covered by "should fail with 404 when retrieve policy"
		It("should fail with 404 when trying to create custom metrics submission without policy", func() {
			_, status := getPolicy()
			Expect(status).To(Equal(404))
		})

		It("should fail to create an invalid custom metrics submission", func() {
			response, status := createPolicy(GenerateBindingsWithScalingPolicy("invalid-value", 1, 2, "memoryused", 30, 100))
			Expect(string(response)).Should(ContainSubstring(`configuration.custom_metrics.metric_submission_strategy.allow_from must be one of the following: \"bound_app\", \"same_app\""}`))
			Expect(status).To(Equal(400))
		})

		It("should succeed to create an valid custom metrics submission", func() {
			By("creating custom metrics submission with 'bound_app'")
			policy := GenerateBindingsWithScalingPolicy("bound_app", 1, 2, "memoryused", 30, 100)
			response, status := createPolicy(policy)
			Expect(string(response)).Should(MatchJSON(policy))
			Expect(status).To(Or(Equal(200), Equal(201)))
			/**
			STEP: PUT 'https://autoscaler-3317.autoscaler.app-runtime-interfaces.ci.cloudfoundry.org/v1/apps/17ae5290-c63a-4836-81a6-42d9635c293a/policy' - /Users/I545443/SAPDevelop/asalan316/app-autoscaler-release/src/acceptance/api/api_test.go:308 @ 11/18/24 13:53:22.373
			  [FAILED] Expected
			      <string>: {
			        "instance_min_count": 1,
			        "instance_max_count": 2,
			        "scaling_rules": [
			          {
			            "metric_type": "memoryused",
			            "breach_duration_secs": 60,
			            "threshold": 100,
			            "operator": "\u003e=",
			            "cool_down_secs": 60,
			            "adjustment": "+1"
			          },
			          {
			            "metric_type": "memoryused",
			            "breach_duration_secs": 60,
			            "threshold": 30,
			            "operator": "\u003c",
			            "cool_down_secs": 60,
			            "adjustment": "-1"
			          }
			        ]
			      }
			  to match JSON of
			      <string>: {
			        "configuration": {
			          "custom_metrics": {
			            "metric_submission_strategy": {
			              "allow_from": ""
			            }
			          }
			        },
			        "instance_min_count": 1,
			        "instance_max_count": 2,
			        "scaling_rules": [
			          {
			            "metric_type": "memoryused",
			            "breach_duration_secs": 60,
			            "threshold": 100,
			            "operator": ">=",
			            "cool_down_secs": 60,
			            "adjustment": "+1"
			          },
			          {
			            "metric_type": "memoryused",
			            "breach_duration_secs": 60,
			            "threshold": 30,
			            "operator": "<",
			            "cool_down_secs": 60,
			            "adjustment": "-1"
			          }
			        ]
			      }


			By("creating custom metrics submission with empty value ' '")
			policy = GenerateBindingsWithScalingPolicy("", 1, 2, "memoryused", 30, 100)
			response, status = createPolicy(policy)
			Expect(string(response)).Should(MatchJSON(policy))
			Expect(status).To(Or(Equal(200), Equal(201)))*/
		})

	})

	When("a scaling policy is set", func() {
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

		When("an unrelated user tries to access the API", func() {
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

		When("a scale out is triggered ", func() {
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

		When("trying to get info for an app not bound to the service", func() {
			BeforeEach(func() {
				UnbindServiceFromApp(cfg, appName, instanceName)
			})
			It("should not be possible to get information from the API", func() {
				By("getting the policy")
				_, status := getPolicy()
				Expect(status).To(Equal(http.StatusForbidden))

				By("getting the history")
				_, status = get(historyURL)
				Expect(status).To(Equal(http.StatusForbidden))

				By("getting the aggregated metrics")
				_, status = get(aggregatedMetricURL)
				Expect(status).To(Equal(http.StatusForbidden))
			})
		})

		Context("and a custom metrics strategy is already set", func() {
			var status int
			var expectedPolicy string
			var actualPolicy []byte
			BeforeEach(func() {
				expectedPolicy = GenerateBindingsWithScalingPolicy("bound_app", 1, 2, "memoryused", 30, 100)
				actualPolicy, status = createPolicy(expectedPolicy)
				Expect(status).To(Equal(200))
			})
			It("should succeed to delete a custom metrics strategy", func() {
				_, status = deletePolicy()
				Expect(status).To(Equal(200))
			})
			It("should succeed to get a custom metrics strategy", func() {
				actualPolicy, status = getPolicy()
				Expect(string(actualPolicy)).Should(MatchJSON(expectedPolicy))
				Expect(status).To(Equal(200))
			})
			It("should fail to update an invalid custom metrics strategy", func() {
				expectedPolicy = GenerateBindingsWithScalingPolicy("invalid-update", 1, 2, "memoryused", 30, 100)
				actualPolicy, status = createPolicy(expectedPolicy)
				Expect(string(actualPolicy)).Should(ContainSubstring(`configuration.custom_metrics.metric_submission_strategy.allow_from must be one of the following: \"bound_app\", \"same_app\""}`))
				Expect(status).To(Equal(400))
			})
			It("should succeed to update a valid custom metrics strategy", func() {

				Expect(string(actualPolicy)).Should(MatchJSON(expectedPolicy))
				Expect(status).To(Equal(200))
			})

			When("custom metrics strategy is removed from the existing policy", func() {
				It("should removed the custom metrics strategy and displays policy only", func() {
					Expect(string(actualPolicy)).Should(MatchJSON(expectedPolicy), "policy and custom metrics strategy should be present")

					By("updating policy without custom metrics strategy")
					expectedPolicy = GenerateDynamicScaleOutPolicy(1, 2, "memoryused", 30)
					actualPolicy, status = createPolicy(expectedPolicy)
					Expect(string(actualPolicy)).Should(MatchJSON(expectedPolicy), "policy should be present only")
					Expect(status).To(Equal(200))
				})
			})
		})
	})
})

func getHistory() *HistoryResults {
	raw, status := get(historyURL)
	Expect(status).To(Equal(200))

	var histories *HistoryResults
	err := json.Unmarshal(raw, &histories)
	Expect(err).ShouldNot(HaveOccurred())
	return histories
}

func getAggregatedMetrics() *AggregatedMetricsResults {
	raw, status := get(aggregatedMetricURL)
	Expect(status).To(Equal(200))
	var metrics *AggregatedMetricsResults
	err := json.Unmarshal(raw, &metrics)
	Expect(err).ShouldNot(HaveOccurred())
	return metrics
}

func createPolicy(policy string) ([]byte, int) {
	return put(policyURL, policy)
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
	return deleteReq(policyURL)
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

func getPolicy() ([]byte, int) {
	return get(policyURL)
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
