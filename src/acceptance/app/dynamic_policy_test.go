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
		policy               string
	)

	BeforeEach(func() {
		instanceName = generator.PrefixedRandomName("autoscaler", "service")
		createService := cf.Cf("create-service", cfg.ServiceName, cfg.ServicePlan, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(createService).To(Exit(0), "failed creating service")
	})

	JustBeforeEach(func() {
		appName = generator.PrefixedRandomName("autoscaler", "nodeapp")
		countStr := strconv.Itoa(initialInstanceCount)
		createApp := cf.Cf("push", appName, "--no-start", "-i", countStr, "-b", cfg.NodejsBuildpackName, "-m", "128M", "-p", config.NODE_APP, "-d", cfg.AppsDomain).Wait(cfg.DefaultTimeoutDuration())
		Expect(createApp).To(Exit(0), "failed creating app")

		guid := cf.Cf("app", appName, "--guid").Wait(cfg.DefaultTimeoutDuration())
		Expect(guid).To(Exit(0))
		appGUID = strings.TrimSpace(string(guid.Out.Contents()))

		Expect(cf.Cf("start", appName).Wait(cfg.CfPushTimeoutDuration())).To(Exit(0))
		waitForNInstancesRunning(appGUID, initialInstanceCount, cfg.DefaultTimeoutDuration())
	})

	AfterEach(func() {
		Expect(cf.Cf("delete", appName, "-f", "-r").Wait(cfg.DefaultTimeoutDuration())).To(Exit(0))

		deleteService := cf.Cf("delete-service", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
		Expect(deleteService).To(Exit(0))
	})

	Context("when scaling by memoryused", func() {

		JustBeforeEach(func() {
			bindService := cf.Cf("bind-service", appName, instanceName, "-c", policy).Wait(cfg.DefaultTimeoutDuration())
			Expect(bindService).To(Exit(0), "failed binding service to app with a policy ")
		})

		AfterEach(func() {
			unbindService := cf.Cf("unbind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
			Expect(unbindService).To(Exit(0), "failed unbinding service from app")
		})

		Context("when memory used is greater than scaling out threshold", func() {
			BeforeEach(func() {
				policy = generateDynamicScaleOutPolicy(1, 2, "memoryused", 30)
				initialInstanceCount = 1
			})

			It("should scale out", func() {
				totalTime := time.Duration(interval*2)*time.Second + 3*time.Minute
				finishTime := time.Now().Add(totalTime)

				Eventually(func() uint64 {
					return averageMemoryUsedByInstance(appGUID, totalTime)
				}, totalTime, 15*time.Second).Should(BeNumerically(">=", 30*MB))

				waitForNInstancesRunning(appGUID, 2, finishTime.Sub(time.Now()))
			})

		})

		Context("when  memory used is lower than scaling in threshold", func() {
			BeforeEach(func() {
				policy = generateDynamicScaleInPolicy(1, 2, "memoryused", 80)
				initialInstanceCount = 2
			})
			It("should scale in", func() {
				totalTime := time.Duration(interval*2)*time.Second + 3*time.Minute
				finishTime := time.Now().Add(totalTime)

				Eventually(func() uint64 {
					return averageMemoryUsedByInstance(appGUID, totalTime)
				}, totalTime, 15*time.Second).Should(BeNumerically("<", 80*MB))

				waitForNInstancesRunning(appGUID, 1, finishTime.Sub(time.Now()))
			})
		})

	})

	Context("when scaling by memoryutil", func() {

		JustBeforeEach(func() {
			bindService := cf.Cf("bind-service", appName, instanceName, "-c", policy).Wait(cfg.DefaultTimeoutDuration())
			Expect(bindService).To(Exit(0), "failed binding service to app with a policy ")
		})

		AfterEach(func() {
			unbindService := cf.Cf("unbind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
			Expect(unbindService).To(Exit(0), "failed unbinding service from app")
		})

		Context("when memoryutil is greater than scaling out threshold", func() {
			BeforeEach(func() {
				policy = generateDynamicScaleOutPolicy(1, 2, "memoryutil", 20)
				initialInstanceCount = 1
			})

			It("should scale out", func() {
				totalTime := time.Duration(interval*2)*time.Second + 3*time.Minute
				finishTime := time.Now().Add(totalTime)

				Eventually(func() uint64 {
					return averageMemoryUsedByInstance(appGUID, totalTime)
				}, totalTime, 15*time.Second).Should(BeNumerically(">=", 26*MB))

				waitForNInstancesRunning(appGUID, 2, finishTime.Sub(time.Now()))
			})

		})

		Context("when  memoryutil is lower than scaling in threshold", func() {
			BeforeEach(func() {
				policy = generateDynamicScaleInPolicy(1, 2, "memoryutil", 80)
				initialInstanceCount = 2
			})
			It("should scale in", func() {
				totalTime := time.Duration(interval*2)*time.Second + 3*time.Minute
				finishTime := time.Now().Add(totalTime)

				Eventually(func() uint64 {
					return averageMemoryUsedByInstance(appGUID, totalTime)
				}, totalTime, 15*time.Second).Should(BeNumerically("<", 115*MB))

				waitForNInstancesRunning(appGUID, 1, finishTime.Sub(time.Now()))
			})
		})

	})
})
