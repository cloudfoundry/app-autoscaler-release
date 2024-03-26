package config_test

import (
	. "acceptance/config"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfigSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ConfigSuite")
}

var _ = Describe("LoadConfig", func() {
	Context("CONFIG env var not set", func() {
		It("terminates suite", func() {
			loadConfigExpectSuiteTerminationWith("Must set $CONFIG to point to a json file")
		})
	})

	Context("CONFIG env var set to non-existing file ", func() {
		BeforeEach(func() {
			err := os.Setenv("CONFIG", "this/path/does/not/exist/config.json")
			Expect(err).ToNot(HaveOccurred())
		})

		It("terminates suite", func() {
			loadConfigExpectSuiteTerminationWith("open this/path/does/not/exist/config.json: no such file or directory")
		})
	})

	Context("CONFIG env var set to existing file", func() {
		var configFile *os.File

		BeforeEach(func() {
			tmpDir := GinkgoT().TempDir()
			tmpFile, err := os.Create(fmt.Sprintf("%s/config.json", tmpDir))
			Expect(err).ShouldNot(HaveOccurred())
			configFile = tmpFile

			err = os.Setenv("CONFIG", configFile.Name())
			Expect(err).ToNot(HaveOccurred())
		})

		DescribeTable("missing required fields", func(json string, message string) {
			write(json, configFile)
			loadConfigExpectSuiteTerminationWith(message)
		},
			Entry("terminates suite because api is missing", `{}`, "missing configuration 'api'"),
			Entry("terminates suite because 'admin_user' is missing", `{
				"api": "api"
			}`, "missing configuration 'admin_user'"),
			Entry("terminates suite because admin_password is missing", `{
				"api": "api",
				"admin_user": "admin_user"
			}`, "missing configuration 'admin_password'"),
			Entry("terminates suite because service_name is missing", `{
				"api": "api",
				"admin_user": "admin_user",
				"admin_password": "admin_password"
			}`, "missing configuration 'service_name'"),
			Entry("terminates suite because service_plan is missing", `{
				"api": "api",
				"admin_user": "admin_user",
				"admin_password": "admin_password",
				"service_name": "service_name"
			}`, "missing configuration 'service_plan'"),
			Entry("terminates suite because aggregate_interval is missing", `{
				"api": "api",
				"admin_user": "admin_user",
				"admin_password": "admin_password",
				"service_name": "service_name",
				"service_plan": "service_plan"
			}`, "missing configuration 'aggregate_interval'"),
			Entry("terminates suite because autoscaler_api is missing", `{
				"api": "api",
				"admin_user": "admin_user",
				"admin_password": "admin_password",
				"service_name": "service_name",
				"service_plan": "service_plan",
				"aggregate_interval": 30
			}`, "missing configuration 'autoscaler_api'"),
		)

		Context("timeout_scale not set correctly", func() {
			It("falls back to a correct value", func() {
				write(configWith(`"timeout_scale": 0`), configFile)
				cfg := LoadConfig(DefaultTerminateSuite)
				Expect(cfg.TimeoutScale).To(Equal(1.0))
			})
		})

		Context("aggregate_interval not set correctly", func() {
			It("falls back to a correct value", func() {
				write(configWith(`"aggregate_interval": 59`), configFile)
				cfg := LoadConfig(DefaultTerminateSuite)
				Expect(cfg.AggregateInterval).To(Equal(60))
			})
		})

		Context("eventgenerator_health_endpoint not set correctly", func() {
			It("falls back to a correct value", func() {
				write(configWith(`"eventgenerator_health_endpoint": "foo.bar/"`), configFile)
				cfg := LoadConfig(DefaultTerminateSuite)
				Expect(cfg.EventgeneratorHealthEndpoint).To(Equal("https://foo.bar"))
			})
		})

		Context("scalingengine_health_endpoint not set correctly", func() {
			It("falls back to a correct value", func() {
				write(configWith(`"scalingengine_health_endpoint": "foo.bar/"`), configFile)
				cfg := LoadConfig(DefaultTerminateSuite)
				Expect(cfg.ScalingengineHealthEndpoint).To(Equal("https://foo.bar"))
			})
		})

		Context("operator_health_endpoint not set correctly", func() {
			It("falls back to a correct value", func() {
				write(configWith(`"operator_health_endpoint": "foo.bar/"`), configFile)
				cfg := LoadConfig(DefaultTerminateSuite)
				Expect(cfg.OperatorHealthEndpoint).To(Equal("https://foo.bar"))
			})
		})

		Context("metricsforwarder_health_endpoint not set correctly", func() {
			It("falls back to a correct value", func() {
				write(configWith(`"metricsforwarder_health_endpoint": "foo.bar/"`), configFile)
				cfg := LoadConfig(DefaultTerminateSuite)
				Expect(cfg.MetricsforwarderHealthEndpoint).To(Equal("https://foo.bar"))
			})
		})

		Context("scheduler_health_endpoint not set correctly", func() {
			It("falls back to a correct value", func() {
				write(configWith(`"scheduler_health_endpoint": "foo.bar/"`), configFile)
				cfg := LoadConfig(DefaultTerminateSuite)
				Expect(cfg.SchedulerHealthEndpoint).To(Equal("https://foo.bar"))
			})
		})

		Context("cpuutil_scaling_policy_test not set", func() {
			It("results in a default value", func() {
				write(config(), configFile)
				cfg := LoadConfig(DefaultTerminateSuite)
				Expect(cfg.CPUUtilScalingPolicyTest.AppMemory).To(Equal("1GB"))
				Expect(cfg.CPUUtilScalingPolicyTest.AppCPUEntitlement).To(Equal(25))
			})
		})
	})
})

func configWith(keyValue string) string {
	// template contains all required fields
	template := `{
		"api": "api",
		"admin_user": "admin_user",
		"admin_password": "admin_password",
		"service_name": "service_name",
		"service_plan": "service_plan",
		"aggregate_interval": 30,
		"autoscaler_api": "autoscaler_api",
		%s
	}`

	return fmt.Sprintf(template, keyValue)
}

func config() string {
	// passing dummy stuff to get a config JSON that comes with all required fields
	return configWith(`"dummyKey": "dummyValue"`)
}

func write(content string, file *os.File) {
	_, err := file.Write([]byte(content))
	Expect(err).ToNot(HaveOccurred())
}

func loadConfigExpectSuiteTerminationWith(expectedMessage string) {
	terminated := false
	actualMessage := ""
	var mockTerminateSuite TerminateSuite = func(message string, _ ...int) {
		terminated = true
		actualMessage = message
		panic(nil)
	}

	defer func() {
		if r := recover(); r == nil {
			Fail("expected a panic to recover from")
		}
		Expect(terminated).To(BeTrue())
		Expect(actualMessage).To(Equal(expectedMessage))
	}()

	LoadConfig(mockTerminateSuite)
}
