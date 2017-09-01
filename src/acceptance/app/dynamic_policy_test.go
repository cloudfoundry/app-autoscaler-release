package app

import (
	"acceptance/config"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
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
		doneChan             chan bool
		ticker               *time.Ticker
	)

	BeforeEach(func() {
		instanceName = generator.PrefixedRandomName("autoscaler", "service")
		createService := cf.Cf("create-service", cfg.ServiceName, cfg.ServicePlan, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(createService).To(Exit(0), "failed creating service")
	})

	JustBeforeEach(func() {
		appName = generator.PrefixedRandomName("autoscaler", "nodeapp")
		countStr := strconv.Itoa(initialInstanceCount)
		createApp := cf.Cf("push", appName, "--no-start", "-i", countStr, "-b", cfg.NodejsBuildpackName, "-m", "128M", "-p", config.NODE_APP, "-d", cfg.AppsDomain).Wait(cfg.CfPushTimeoutDuration())
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

	Context("when scaling by responsetime", func() {

		JustBeforeEach(func() {
			bindService := cf.Cf("bind-service", appName, instanceName, "-c", policy).Wait(cfg.DefaultTimeoutDuration())
			Expect(bindService).To(Exit(0), "failed binding service to app with a policy ")
			doneChan = make(chan bool)
		})

		AfterEach(func() {
			doneChan <- true
			unbindService := cf.Cf("unbind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
			Expect(unbindService).To(Exit(0), "failed unbinding service from app")
		})

		Context("when responsetime is greater than scaling out threshold", func() {

			BeforeEach(func() {
				policy = generateDynamicScaleOutPolicy(1, 2, "responsetime", 3000)
				initialInstanceCount = 1
			})

			JustBeforeEach(func() {
				ticker = time.NewTicker(10 * time.Second)
				go func(chan bool) {
					defer GinkgoRecover()
					for {
						select {
						case <-doneChan:
							ticker.Stop()
							return
						case <-ticker.C:
							Eventually(func() string {
								return helpers.CurlAppWithTimeout(cfg, appName, "/slow/10000", 1*time.Minute)
							}, 1*time.Minute, 2*time.Second).Should(ContainSubstring("dummy application with slow response"))
						}
					}
				}(doneChan)
			})

			It("should scale out", func() {
				finishTime := time.Duration(interval*2)*time.Second + 5*time.Minute
				waitForNInstancesRunning(appGUID, 2, finishTime)
			})
		})

		Context("when responsetime is less than scaling in threshold", func() {

			BeforeEach(func() {
				policy = generateDynamicScaleInPolicy(1, 2, "responsetime", 1000)
				initialInstanceCount = 2
			})

			JustBeforeEach(func() {
				ticker = time.NewTicker(2 * time.Second)
				go func(chan bool) {
					defer GinkgoRecover()
					for {
						select {
						case <-doneChan:
							ticker.Stop()
							return
						case <-ticker.C:
							Eventually(func() string {
								return helpers.CurlAppWithTimeout(cfg, appName, "/fast", 10*time.Second)
							}, 10*time.Second, 1*time.Second).Should(ContainSubstring("dummy application with fast response"))
						}
					}
				}(doneChan)
			})

			It("should scale in", func() {
				finishTime := time.Duration(interval*2)*time.Second + 5*time.Minute
				waitForNInstancesRunning(appGUID, 1, finishTime)
			})
		})

	})

	Context("when scaling by throughput", func() {

		JustBeforeEach(func() {
			bindService := cf.Cf("bind-service", appName, instanceName, "-c", policy).Wait(cfg.DefaultTimeoutDuration())
			Expect(bindService).To(Exit(0), "failed binding service to app with a policy ")
			doneChan = make(chan bool)
		})

		AfterEach(func() {
			doneChan <- true
			unbindService := cf.Cf("unbind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
			Expect(unbindService).To(Exit(0), "failed unbinding service from app")
		})

		Context("when throughput is greater than scaling out threshold", func() {

			BeforeEach(func() {
				policy = generateDynamicScaleOutPolicy(1, 2, "throughput", 2)
				initialInstanceCount = 1
			})

			JustBeforeEach(func() {
				ticker = time.NewTicker(25 * time.Millisecond)
				go func(chan bool) {
					defer GinkgoRecover()
					for {
						select {
						case <-doneChan:
							ticker.Stop()
							return
						case <-ticker.C:
							Eventually(func() string {
								return helpers.CurlAppWithTimeout(cfg, appName, "/fast", 10*time.Second)
							}, 10*time.Second, 25*time.Millisecond).Should(ContainSubstring("dummy application with fast response"))
						}
					}
				}(doneChan)
			})

			It("should scale out", func() {
				finishTime := time.Duration(interval*2)*time.Second + 3*time.Minute
				waitForNInstancesRunning(appGUID, 2, finishTime)
			})

		})

		Context("when throughput is less than scaling in threshold", func() {

			BeforeEach(func() {
				policy = generateDynamicScaleInPolicy(1, 2, "throughput", 1)
				initialInstanceCount = 2
			})

			JustBeforeEach(func() {
				ticker = time.NewTicker(10 * time.Second)
				go func(chan bool) {
					defer GinkgoRecover()
					for {
						select {
						case <-doneChan:
							ticker.Stop()
							return
						case <-ticker.C:
							Eventually(func() string {
								return helpers.CurlAppWithTimeout(cfg, appName, "/fast", 10*time.Second)
							}, 10*time.Second, 1*time.Second).Should(ContainSubstring("dummy application with fast response"))
						}
					}
				}(doneChan)
			})
			It("should scale in", func() {
				finishTime := time.Duration(interval*2)*time.Second + 3*time.Minute
				waitForNInstancesRunning(appGUID, 1, finishTime)
			})
		})

	})

})
