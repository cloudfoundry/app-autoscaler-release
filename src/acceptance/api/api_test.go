package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	. "acceptance/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

	var (
		policy string
		body   io.Reader
	)

	BeforeEach(func() {
		oauthToken = OauthToken(cfg)
	})

	It("should succeed to check health", func() {
		req, err := http.NewRequest("GET", healthURL, nil)
		resp, err := DoAPIRequest(req)
		Expect(err).ShouldNot(HaveOccurred())

		defer resp.Body.Close()

		Expect(err).ShouldNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(200))
	})

	Context("when no policy defined", func() {

		BeforeEach(func() {
			//delete policy here to make sure the condtion "no policy defined"
			req, err := http.NewRequest("DELETE", policyURL, nil)
			req.Header.Add("Authorization", oauthToken)
			req.Header.Add("Content-Type", "application/json")
			resp, err := DoAPIRequest(req)
			Expect(err).ShouldNot(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Or(Equal(200), Equal(404)))
		})

		It("should fail with 404 when retrieve policy", func() {
			req, err := http.NewRequest("GET", policyURL, nil)
			req.Header.Add("Authorization", oauthToken)
			req.Header.Add("Content-Type", "application/json")

			resp, err := DoAPIRequest(req)
			Expect(err).ShouldNot(HaveOccurred())

			defer resp.Body.Close()

			_, err = ioutil.ReadAll(resp.Body)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(404))

		})

		It("should succeed to create a valid policy", func() {
			policy = GenerateDynamicScaleOutPolicy(cfg, 1, 2, "memoryused", 30)
			body = bytes.NewBuffer([]byte(policy))

			req, err := http.NewRequest("PUT", policyURL, body)
			req.Header.Add("Authorization", oauthToken)
			req.Header.Add("Content-Type", "application/json")

			resp, err := DoAPIRequest(req)
			Expect(err).ShouldNot(HaveOccurred())

			defer resp.Body.Close()

			raw, err := ioutil.ReadAll(resp.Body)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).To(Or(Equal(200), Equal(201)))
			var responsedPolicy *ScalingPolicy
			err = json.Unmarshal(raw, &responsedPolicy)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(string(raw)).Should(Equal(strings.TrimSpace(policy)))

		})

		It("should fail to create an invalid policy", func() {
			policy = GenerateDynamicScaleOutPolicy(cfg, 0, 2, "memoryused", 30)
			body = bytes.NewBuffer([]byte(policy))

			req, err := http.NewRequest("PUT", policyURL, body)
			req.Header.Add("Authorization", oauthToken)
			req.Header.Add("Content-Type", "application/json")

			resp, err := DoAPIRequest(req)
			Expect(err).ShouldNot(HaveOccurred())

			defer resp.Body.Close()

			raw, err := ioutil.ReadAll(resp.Body)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(400))
			Expect(string(raw)).Should(ContainSubstring("instance.instance_min_count must have a minimum value of 1"))
		})

	})

	Context("When policy is defined", func() {

		BeforeEach(func() {
			policy = GenerateDynamicScaleOutPolicy(cfg, 1, 2, "memoryused", 30)
			body = bytes.NewBuffer([]byte(policy))

			req, err := http.NewRequest("PUT", policyURL, body)
			req.Header.Add("Authorization", oauthToken)
			req.Header.Add("Content-Type", "application/json")

			resp, err := DoAPIRequest(req)
			Expect(err).ShouldNot(HaveOccurred())

			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Or(Equal(200), Equal(201)))
		})

		It("should succeed to delete a policy", func() {
			req, err := http.NewRequest("DELETE", policyURL, nil)
			req.Header.Add("Authorization", oauthToken)
			req.Header.Add("Content-Type", "application/json")

			resp, err := DoAPIRequest(req)
			Expect(err).ShouldNot(HaveOccurred())

			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(200))
		})

		It("should succeed to get a policy", func() {

			req, err := http.NewRequest("GET", policyURL, nil)
			req.Header.Add("Authorization", oauthToken)
			req.Header.Add("Content-Type", "application/json")

			resp, err := DoAPIRequest(req)
			Expect(err).ShouldNot(HaveOccurred())

			defer resp.Body.Close()

			raw, err := ioutil.ReadAll(resp.Body)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			var responsedPolicy *ScalingPolicy
			err = json.Unmarshal(raw, &responsedPolicy)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(string(raw)).Should(Equal(strings.TrimSpace(policy)))
		})

		It("should succeed to update a valid policy", func() {
			newpolicy := GenerateDynamicScaleOutPolicy(cfg, 1, 2, "memoryused", 30)
			body = bytes.NewBuffer([]byte(newpolicy))

			req, err := http.NewRequest("PUT", policyURL, body)
			req.Header.Add("Authorization", oauthToken)
			req.Header.Add("Content-Type", "application/json")

			resp, err := DoAPIRequest(req)
			Expect(err).ShouldNot(HaveOccurred())

			defer resp.Body.Close()

			raw, err := ioutil.ReadAll(resp.Body)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			var responsedPolicy *ScalingPolicy
			err = json.Unmarshal(raw, &responsedPolicy)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(string(raw)).Should(Equal(strings.TrimSpace(newpolicy)))

		})

		It("should fail to update an invalid policy", func() {
			By("return 400 when the new policy is invalid")
			newpolicy := GenerateDynamicScaleOutPolicy(cfg, 0, 2, "memoryused", 30)
			body = bytes.NewBuffer([]byte(newpolicy))

			req, err := http.NewRequest("PUT", policyURL, body)
			req.Header.Add("Authorization", oauthToken)
			req.Header.Add("Content-Type", "application/json")

			resp, err := DoAPIRequest(req)
			Expect(err).ShouldNot(HaveOccurred())

			defer resp.Body.Close()

			raw, err := ioutil.ReadAll(resp.Body)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(400))

			By("the original policy is not changed")
			req, err = http.NewRequest("GET", policyURL, nil)
			req.Header.Add("Authorization", oauthToken)
			req.Header.Add("Content-Type", "application/json")

			resp, err = DoAPIRequest(req)
			Expect(err).ShouldNot(HaveOccurred())

			defer resp.Body.Close()

			raw, err = ioutil.ReadAll(resp.Body)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			var responsedPolicy *ScalingPolicy
			err = json.Unmarshal(raw, &responsedPolicy)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(string(raw)).Should(Equal(strings.TrimSpace(policy)))

		})

		Context("When scale out is triggered ", func() {

			BeforeEach(func() {
				totalTime := time.Duration(cfg.AggregateInterval*2)*time.Second + 3*time.Minute
				finishTime := time.Now().Add(totalTime)

				Eventually(func() uint64 {
					return AverageMemoryUsedByInstance(appGUID, totalTime)
				}, totalTime, 15*time.Second).Should(BeNumerically(">=", 30*MB))

				WaitForNInstancesRunning(appGUID, 2, finishTime.Sub(time.Now()))
			})

			It("should succeed to get instance metrics", func() {

				req, err := http.NewRequest("GET", metricURL, nil)
				req.Header.Add("Authorization", oauthToken)
				req.Header.Add("Content-Type", "application/json")

				resp, err := DoAPIRequest(req)
				Expect(err).ShouldNot(HaveOccurred())

				defer resp.Body.Close()

				raw, err := ioutil.ReadAll(resp.Body)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(200))

				var metrics *MetricsResults
				err = json.Unmarshal(raw, &metrics)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(metrics.Metrics)).Should(BeNumerically(">=", 1))
			})

			It("should succeed to get aggregated metrics", func() {

				req, err := http.NewRequest("GET", aggregatedMetricURL, nil)
				req.Header.Add("Authorization", oauthToken)
				req.Header.Add("Content-Type", "application/json")

				resp, err := DoAPIRequest(req)
				Expect(err).ShouldNot(HaveOccurred())

				defer resp.Body.Close()

				raw, err := ioutil.ReadAll(resp.Body)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(200))

				var metrics *AggregatedMetricsResults
				err = json.Unmarshal(raw, &metrics)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(metrics.Metrics)).Should(BeNumerically(">=", 1))
			})

			It("should succeed to get histories", func() {
				req, err := http.NewRequest("GET", historyURL, nil)
				req.Header.Add("Authorization", oauthToken)
				req.Header.Add("Content-Type", "application/json")

				resp, err := DoAPIRequest(req)
				Expect(err).ShouldNot(HaveOccurred())

				defer resp.Body.Close()

				raw, err := ioutil.ReadAll(resp.Body)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(200))

				var histories *HistoryResults
				err = json.Unmarshal(raw, &histories)
				Expect(err).ShouldNot(HaveOccurred())

				for _, entry := range histories.Histories {
					Expect(entry.AppId).To(Equal(appGUID))
					Expect(entry.ScalingType).Should(BeNumerically("==", 0))
					Expect(entry.Status).Should(BeNumerically("==", 0))
					Expect(entry.Reason).To(Equal(fmt.Sprintf("+1 instance(s) because memoryused >= 30MB for %d seconds", TestBreachDurationSeconds)))
				}

			})
		})
	})

})
