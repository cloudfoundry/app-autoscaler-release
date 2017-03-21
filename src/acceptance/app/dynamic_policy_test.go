package app

import (
	"acceptance/config"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"strconv"
	"strings"
	"time"
)

var _ = Describe("AutoScaler dynamic policy", func() {
	var (
		appName              string
		appGUID              string
		instanceName         string
		initialInstanceCount int
		policyFileName       string
	)

	BeforeEach(func() {
		instanceName = generator.PrefixedRandomName("autoscaler", "service")
		createService := cf.Cf("create-service", cfg.ServiceName, cfg.ServicePlan, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(createService).To(Exit(0), "failed creating service")
	})

	JustBeforeEach(func() {
		appName = generator.PrefixedRandomName("autoscaler", "nodeapp")
		countStr := strconv.Itoa(initialInstanceCount)
		createApp := cf.Cf("push", appName, "--no-start", "-i", countStr, "-b", cfg.NodejsBuildpackName, "-m", cfg.NodeMemoryLimit, "-p", config.NODE_APP, "-d", cfg.AppsDomain).Wait(cfg.DefaultTimeoutDuration())
		Expect(createApp).To(Exit(0), "failed creating app")

		guid := cf.Cf("app", appName, "--guid").Wait(cfg.DefaultTimeout)
		Expect(guid).To(Exit(0))
		appGUID = strings.TrimSpace(string(guid.Out.Contents()))

		Expect(cf.Cf("start", appName).Wait(cfg.DefaultTimeout * 2)).To(Exit(0))
		waitForNInstancesRunning(appGUID, initialInstanceCount, cfg.DefaultTimeoutDuration())
	})

	AfterEach(func() {
		Expect(cf.Cf("delete", appName, "-f", "-r").Wait(cfg.CfPushTimeoutDuration())).To(Exit(0))

		deleteService := cf.Cf("delete-service", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
		Expect(deleteService).To(Exit(0))
	})

	Context("when scale by memoryused", func() {

		JustBeforeEach(func() {
			bindService := cf.Cf("bind-service", appName, instanceName, "-c", policyFileName).Wait(cfg.DefaultTimeoutDuration())
			Expect(bindService).To(Exit(0), "failed binding service to app with a policy ")
		})

		AfterEach(func() {
			unbindService := cf.Cf("unbind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
			Expect(unbindService).To(Exit(0), "failed unbinding service from app")
		})

		Context("and 1 instance initially", func() {
			BeforeEach(func() {
				initialInstanceCount = 1
				policyFileName = "../assets/file/policy/dynamic_scale_out.json"
			})

			It("should scale out", func() {
				totalTime := time.Duration(cfg.ReportInterval*2)*time.Second + 2*time.Minute
				finishTime := time.Now().Add(totalTime)

				// make sure our threshold is >= 30 MB
				Eventually(func() uint64 {
					return averageMemoryUsedByInstance(appGUID, totalTime)
				}, totalTime, 15*time.Second).Should(BeNumerically(">=", 30*MB))

				waitForNInstancesRunning(appGUID, initialInstanceCount+1, finishTime.Sub(time.Now()))
			})

		})

		Context("and 2 instances initially", func() {
			BeforeEach(func() {
				initialInstanceCount = 2
				policyFileName = "../assets/file/policy/dynamic_scale_in.json"
			})

			It("should scale in", func() {
				totalTime := time.Duration(cfg.ReportInterval*2)*time.Second + time.Minute
				finishTime := time.Now().Add(totalTime)

				// make sure our threshold is < 60 MB
				Eventually(func() uint64 {
					return averageMemoryUsedByInstance(appGUID, totalTime)
				}, totalTime, 15*time.Second).Should(BeNumerically("<", 60*MB))

				waitForNInstancesRunning(appGUID, initialInstanceCount-1, finishTime.Sub(time.Now()))
			})
		})

	})
})
