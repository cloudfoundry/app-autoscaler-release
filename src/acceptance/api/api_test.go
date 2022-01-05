package api_test

import (
	"fmt"
	"time"

	. "acceptance/helpers"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var (
	oauthToken string
)

var _ = Describe("AutoScaler Public API", func() {

	BeforeEach(func() {
		oauthToken = OauthToken(cfg)
	})

	Context("when no policy defined", func() {

		BeforeEach(func() {
			_, status := app.DeletePolicy()
			Expect(status).To(Or(Equal(200), Equal(404)))
		})

		It("should fail with 404 when retrieve policy", func() {
			_, status := app.GetPolicy()
			Expect(status).To(Equal(404))
		})

		It("should succeed to create a valid policy", func() {
			policy := GenerateDynamicScaleOutPolicy(1, 2, "memoryused", 30)
			newPolicy, status := app.CreatePolicy(policy)
			Expect(status).To(Or(Equal(200), Equal(201)))
			Expect(string(newPolicy)).Should(MatchJSON(policy))
		})

		It("should fail to create an invalid policy", func() {
			response, status := app.CreatePolicy(GenerateDynamicScaleOutPolicy(0, 2, "memoryused", 30))
			Expect(status).To(Equal(400))
			Expect(string(response)).Should(ContainSubstring(`[{"context":"(root).instance_min_count","description":"Must be greater than or equal to 1"}]`))
		})

	})

	Context("When policy is defined", func() {
		memThreshold := int64(10)
		var policy string

		BeforeEach(func() {
			policy = GenerateDynamicScaleOutPolicy(1, 2, "memoryused", memThreshold)
			_, status := app.CreatePolicy(policy)
			Expect(status).To(Or(Equal(200), Equal(201)))
		})

		It("should succeed to delete a policy", func() {
			_, status := app.DeletePolicy()
			Expect(status).To(Equal(200))
		})

		It("should succeed to get a policy", func() {
			gotPolicy, status := app.GetPolicy()
			Expect(status).To(Equal(200))
			Expect(string(gotPolicy)).Should(MatchJSON(policy))
		})

		It("should succeed to update a valid policy", func() {
			newPolicy, status := app.CreatePolicy(GenerateDynamicScaleOutPolicy(1, 2, "memoryused", memThreshold))
			Expect(status).To(Equal(200))
			Expect(string(newPolicy)).Should(MatchJSON(policy))
		})

		It("should fail to update an invalid policy", func() {
			By("return 400 when the new policy is invalid")
			_, status := app.CreatePolicy(GenerateDynamicScaleOutPolicy(0, 2, "memoryused", 30))
			Expect(status).To(Equal(400))

			By("the original policy is not changed")
			existing, status := app.GetPolicy()
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
				workflowhelpers.AsUser(otherSetup.RegularUserContext(), cfg.DefaultTimeoutDuration(), func() { oauthToken = OauthToken(cfg) })
			})

			It("should not be possible to read the policy", func() {
				_, status := app.GetPolicy()
				Expect(status).To(Equal(401))
			})
		})

		Context("When scale out is triggered ", func() {

			BeforeEach(func() {
				totalTime := time.Duration(cfg.AggregateInterval*2)*time.Second + 3*time.Minute

				Eventually(func() uint64 {
					return AverageMemoryUsedByInstance(app.GUID, totalTime)
				}, totalTime, 15*time.Second).Should(BeNumerically(">=", memThreshold*MB))

				WaitForNInstancesRunning(app.GUID, 2, totalTime)
			})

			It("should successfully scale out", func() {

				By("check instance metrics")
				Expect(len(app.Metrics().Metrics)).Should(BeNumerically(">=", 1))

				By("check aggregated metrics")
				Expect(len(app.AggregatedMetrics().Metrics)).Should(BeNumerically(">=", 1))

				By("check history has scale event")
				for _, entry := range app.History().Histories {
					Expect(entry.AppId).To(Equal(app.GUID))
					Expect(entry.ScalingType).Should(BeNumerically("==", 0))
					Expect(entry.Status).Should(BeNumerically("==", 0))
					Expect(entry.Reason).To(Equal(fmt.Sprintf("+1 instance(s) because memoryused >= %dMB for %d seconds", memThreshold, TestBreachDurationSeconds)))
				}
			})
		})
	})
})
