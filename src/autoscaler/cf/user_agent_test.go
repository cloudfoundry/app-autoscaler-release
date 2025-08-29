package cf_test

import (
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/cf"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"runtime"
	"strings"
)

var _ = Describe("User Agent", func() {
	Context("GetUserAgent", func() {
		It("returns a valid user agent string", func() {
			userAgent := cf.GetUserAgent()

			Expect(userAgent).To(ContainSubstring("app-autoscaler/"))
			Expect(userAgent).To(ContainSubstring("Go/" + runtime.Version()))
			Expect(userAgent).To(ContainSubstring(runtime.GOOS + "/" + runtime.GOARCH))
		})

		It("has the expected format", func() {
			userAgent := cf.GetUserAgent()

			// Expected format: app-autoscaler/{version} ({gitRepo}; {commitId}) Go/{goVersion} {os}/{arch}
			parts := strings.Split(userAgent, " ")
			Expect(len(parts)).To(BeNumerically(">=", 3))

			// Check product/version part
			productPart := parts[0]
			Expect(productPart).To(HavePrefix("app-autoscaler/"))

			// Check system info part is wrapped in parentheses
			systemInfoStart := strings.Index(userAgent, "(")
			systemInfoEnd := strings.Index(userAgent, ")")
			Expect(systemInfoStart).To(BeNumerically(">", 0))
			Expect(systemInfoEnd).To(BeNumerically(">", systemInfoStart))

			systemInfo := userAgent[systemInfoStart+1 : systemInfoEnd]
			Expect(systemInfo).To(ContainSubstring(";"))
		})

		It("includes build information when available", func() {
			userAgent := cf.GetUserAgent()

			// The user agent should contain some form of version/build info
			// Even if it's "unknown" or "dev", it should be present
			Expect(userAgent).ToNot(BeEmpty())
			Expect(userAgent).To(MatchRegexp(`app-autoscaler/\w+`))
		})

		It("includes platform information", func() {
			userAgent := cf.GetUserAgent()

			// Should include Go version and platform info
			Expect(userAgent).To(ContainSubstring("Go/"))
			Expect(userAgent).To(MatchRegexp(`\w+/\w+$`)) // os/arch at the end
		})
	})
})
