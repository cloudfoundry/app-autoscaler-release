package app

import (
	"acceptance/config"
	. "acceptance/helpers"

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
		appName string
		appGUID string
		policy  string

		doneChan       chan bool
		doneAcceptChan chan bool
		ticker         *time.Ticker
	)

	BeforeEach(func() {

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
		WaitForNInstancesRunning(appGUID, initialInstanceCount, cfg.DefaultTimeoutDuration())
		CreatePolicy(appName, appGUID, policy)

	})

	AfterEach(func() {
		DeletePolicy(appName, appGUID)
		Expect(cf.Cf("delete", appName, "-f", "-r").Wait(cfg.DefaultTimeoutDuration())).To(Exit(0))
	})

	Context("when scaling by memoryused", func() {

		Context("when memory used is greater than scaling out threshold", func() {
			BeforeEach(func() {
				policy = GenerateDynamicScaleOutPolicy(cfg, 1, 2, "memoryused", 30)
				initialInstanceCount = 1
			})

			It("should scale out", func() {
				totalTime := time.Duration(interval*2)*time.Second + 3*time.Minute
				finishTime := time.Now().Add(totalTime)

				Eventually(func() uint64 {
					return AverageMemoryUsedByInstance(appGUID, totalTime)
				}, totalTime, 15*time.Second).Should(BeNumerically(">=", 30*MB))

				WaitForNInstancesRunning(appGUID, 2, finishTime.Sub(time.Now()))
			})

		})

		Context("when  memory used is lower than scaling in threshold", func() {
			BeforeEach(func() {
				policy = GenerateDynamicScaleInPolicy(cfg, 1, 2, "memoryused", 80)
				initialInstanceCount = 2
			})
			It("should scale in", func() {
				totalTime := time.Duration(interval*2)*time.Second + 3*time.Minute
				finishTime := time.Now().Add(totalTime)

				Eventually(func() uint64 {
					return AverageMemoryUsedByInstance(appGUID, totalTime)
				}, totalTime, 15*time.Second).Should(BeNumerically("<", 80*MB))

				WaitForNInstancesRunning(appGUID, 1, finishTime.Sub(time.Now()))
			})
		})

	})

	Context("when scaling by memoryutil", func() {

		Context("when memoryutil is greater than scaling out threshold", func() {
			BeforeEach(func() {
				policy = GenerateDynamicScaleOutPolicy(cfg, 1, 2, "memoryutil", 20)
				initialInstanceCount = 1
			})

			It("should scale out", func() {
				totalTime := time.Duration(interval*2)*time.Second + 3*time.Minute
				finishTime := time.Now().Add(totalTime)

				Eventually(func() uint64 {
					return AverageMemoryUsedByInstance(appGUID, totalTime)
				}, totalTime, 15*time.Second).Should(BeNumerically(">=", 26*MB))

				WaitForNInstancesRunning(appGUID, 2, finishTime.Sub(time.Now()))
			})

		})

		Context("when  memoryutil is lower than scaling in threshold", func() {
			BeforeEach(func() {
				policy = GenerateDynamicScaleInPolicy(cfg, 1, 2, "memoryutil", 80)
				initialInstanceCount = 2
			})
			It("should scale in", func() {
				totalTime := time.Duration(interval*2)*time.Second + 3*time.Minute
				finishTime := time.Now().Add(totalTime)

				Eventually(func() uint64 {
					return AverageMemoryUsedByInstance(appGUID, totalTime)
				}, totalTime, 15*time.Second).Should(BeNumerically("<", 115*MB))

				WaitForNInstancesRunning(appGUID, 1, finishTime.Sub(time.Now()))
			})
		})

	})

	Context("when scaling by responsetime", func() {

		JustBeforeEach(func() {
			doneChan = make(chan bool)
			doneAcceptChan = make(chan bool)
		})

		AfterEach(func() {
			close(doneChan)
			Eventually(doneAcceptChan, 10*time.Second).Should(Receive())
		})

		Context("when responsetime is greater than scaling out threshold", func() {

			BeforeEach(func() {
				policy = GenerateDynamicScaleOutPolicy(cfg, 1, 2, "responsetime", 2000)
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
							doneAcceptChan <- true
							return
						case <-ticker.C:
							Eventually(func() string {
								return helpers.CurlAppWithTimeout(cfg, appName, "/slow/3000", 10*time.Second)
							}, 10*time.Second, 1*time.Second).Should(ContainSubstring("dummy application with slow response"))
						}
					}
				}(doneChan)
			})

			It("should scale out", func() {
				finishTime := time.Duration(interval*2)*time.Second + 5*time.Minute
				WaitForNInstancesRunning(appGUID, 2, finishTime)
			})
		})

		Context("when responsetime is less than scaling in threshold", func() {

			BeforeEach(func() {
				policy = GenerateDynamicScaleInPolicy(cfg, 1, 2, "responsetime", 1000)
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
							doneAcceptChan <- true
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
				WaitForNInstancesRunning(appGUID, 1, finishTime)
			})
		})

	})

	Context("when scaling by throughput", func() {

		JustBeforeEach(func() {
			doneChan = make(chan bool)
			doneAcceptChan = make(chan bool)
		})

		AfterEach(func() {
			close(doneChan)
			Eventually(doneAcceptChan, 10*time.Second).Should(Receive())
		})

		Context("when throughput is greater than scaling out threshold", func() {

			BeforeEach(func() {
				policy = GenerateDynamicScaleOutPolicy(cfg, 1, 2, "throughput", 1)
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
							doneAcceptChan <- true
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
				WaitForNInstancesRunning(appGUID, 2, finishTime)
			})

		})

		Context("when throughput is less than scaling in threshold", func() {

			BeforeEach(func() {
				policy = GenerateDynamicScaleInPolicy(cfg, 1, 2, "throughput", 1)
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
							doneAcceptChan <- true
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
				WaitForNInstancesRunning(appGUID, 1, finishTime)
			})
		})

	})

})
