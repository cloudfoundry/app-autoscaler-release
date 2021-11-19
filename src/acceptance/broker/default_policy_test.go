package broker_test

import (
	. "github.com/onsi/ginkgo"
)

var _ = XDescribe("Default policy tests", func() {
	//var appName string
	//var instance * broker.Instance

	BeforeEach(func() {
		//appName = helpers.CreateTestApp(cfg, "default-policy", 1)
	})

	When("a default policy is available in the service instance", func() {
		BeforeEach(func() {
			// Create service instance with default policy
			//instance = broker.CreateInstance(cfg)
			// TODO: Modify the createService function to allow default plan setting, change the implementation so that create-service allows default_policy parameter.
		})

		It("All app bindings with empty policies will inherit the default policy", func() {
		})

		It("Should use binding policy over default policy", func() {
		})

		When("updating service default policy ", func() {
			It("Should update all bindings using the default policy", func() {
			})
		})

		Describe("When removing existing binding policy", func() {
			It("Should use the default policy instead", func() {
			})
		})
	})
})
