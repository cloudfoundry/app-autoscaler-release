package pre_upgrade_test

import (
	. "acceptance/helpers"

	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"strconv"
	"time"
)

var _ = Describe("AutoScaler dynamic policy", func() {
	var (
		appName string
		appGUID string
		policy  string

		initialInstanceCount = 1
	)

	JustBeforeEach(func() {
		appName = generator.PrefixedRandomName("autoscaler", "nodeapp-cpu")
		countStr := strconv.Itoa(initialInstanceCount)
		createApp := cf.Cf("push", appName, "--no-start", "--no-route", "-i", countStr, "-b", cfg.NodejsBuildpackName, "-m", "128M", "-p", "../../assets/app/nodeApp").Wait(cfg.CfPushTimeoutDuration())
		Expect(createApp).To(Exit(0), "failed creating app")

		mapRouteToApp := cf.Cf("map-route", appName, cfg.AppsDomain, "--hostname", appName).Wait(cfg.DefaultTimeoutDuration())
		Expect(mapRouteToApp).To(Exit(0), "failed to map route to app")

		appGUID = GetAppGuid(cfg, appName)
		Expect(cf.Cf("start", appName).Wait(cfg.CfPushTimeoutDuration())).To(Exit(0))
		WaitForNInstancesRunning(appGUID, initialInstanceCount, cfg.DefaultTimeoutDuration())
		_ = CreatePolicy(cfg, appName, appGUID, policy)
	})

	Context("when scaling by cpu", func() {

		Context("when cpu is greater than and then less than threshold", func() {

			BeforeEach(func() {
				policy = GenerateDynamicScaleOutAndInPolicy(1, 2, "cpu", 2, 10)
				initialInstanceCount = 1
			})

			It("should scale out and back in", func() {
				response := helpers.CurlAppWithTimeout(cfg, appName, "/cpu/50/1", 10*time.Second)
				Expect(response).Should(ContainSubstring(`set app cpu utilization to 50% for 1 minutes, busyTime=10, idleTime=10`))

				totalTime := time.Duration(interval*2)*time.Second + 3*time.Minute
				finishTime := time.Now().Add(totalTime)

				Eventually(func() float64 {
					return AverageCPUByInstance(appGUID, totalTime)
				}, totalTime, 15*time.Second).Should(BeNumerically(">=", 0.02))

				WaitForNInstancesRunning(appGUID, 2, time.Until(finishTime))

				// lets attempt to scale back down
				response = helpers.CurlAppWithTimeout(cfg, appName, "/cpu/0/5", 10*time.Second)
				Expect(response).Should(ContainSubstring(`set app cpu utilization to 1% for 5 minutes, busyTime=10, idleTime=990`))

				totalTime = time.Duration(interval*2)*time.Second + 10*time.Minute
				finishTime = time.Now().Add(totalTime)

				Eventually(func() float64 {
					return AverageCPUByInstance(appGUID, totalTime)
				}, totalTime, 15*time.Second).Should(BeNumerically("<=", 0.1))

				WaitForNInstancesRunning(appGUID, 1, time.Until(finishTime))
			})

		})

	})

})
