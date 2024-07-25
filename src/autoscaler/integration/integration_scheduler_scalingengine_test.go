package integration_test

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/testhelpers"
	"github.com/google/uuid"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Integration_Scheduler_ScalingEngine", func() {
	var (
		testAppId         string
		testGuid          string
		initInstanceCount = 2
		policyStr         string

		schedulerURL     url.URL
		scalingEngineURL url.URL
	)

	BeforeEach(func() {
		httpClientForScheduler = testhelpers.NewSchedulerClient()

		testAppId = uuid.NewString()
		testGuid = uuid.NewString()
		startFakeCCNOAAUAA(initInstanceCount)

		scalingEngineConfPath = components.PrepareScalingEngineConfig(dbUrl, components.Ports[ScalingEngine], fakeCCNOAAUAA.URL(), defaultHttpClientTimeout, tmpDir)
		startScalingEngine()

		schedulerURL = url.URL{
			Scheme: "https",
			Host:   fmt.Sprintf("127.0.0.1:%d", components.Ports[Scheduler]),
		}

		scalingEngineURL = url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("127.0.0.1:%d", components.Ports[ScalingEngine]),
		}

		schedulerConfPath = components.PrepareSchedulerConfig(dbUrl, scalingEngineURL, tmpDir, defaultHttpClientTimeout)
		schedulerConfPath = components.PrepareSchedulerConfig(dbUrl, scalingEngineURL, tmpDir, defaultHttpClientTimeout)
		startScheduler()

		policyStr = setPolicySpecificDateTime(readPolicyFromFile("fakePolicyWithSpecificDateSchedule.json"), 70*time.Second, 2*time.Hour)

	})

	AfterEach(func() {
		deletePolicy(testAppId)
		stopScheduler()
		stopScalingEngine()
	})

	Describe("Create Schedule", func() {
		Context("Valid specific date schedule", func() {

			AfterEach(func() {
				_, err := deleteSchedule(schedulerURL, testAppId)
				Expect(err).NotTo(HaveOccurred())
			})

			It("creates active schedule in scaling engine", func() {
				resp, err := createSchedule(testAppId, testGuid, policyStr, schedulerURL)
				checkResponseEmptyAndStatusCode(resp, err, http.StatusOK)

				Eventually(func() bool {
					return activeScheduleExists(scalingEngineURL, testAppId)
				}, 2*time.Minute, 1*time.Second).Should(BeTrue())

			})
		})

	})

	Describe("Delete Schedule", func() {
		BeforeEach(func() {
			resp, err := createSchedule(testAppId, testGuid, policyStr, schedulerURL)
			checkResponseEmptyAndStatusCode(resp, err, http.StatusOK)

			Eventually(func() bool {
				return activeScheduleExists(scalingEngineURL, testAppId)
			}, 2*time.Minute, 1*time.Second).Should(BeTrue())
		})

		It("deletes active schedule in scaling engine", func() {
			resp, err := deleteSchedule(schedulerURL, testAppId)
			checkResponseEmptyAndStatusCode(resp, err, http.StatusNoContent)

			Eventually(func() bool {
				return activeScheduleExists(scalingEngineURL, testAppId)
			}, 2*time.Minute, 1*time.Second).Should(BeFalse())
		})
	})

})
