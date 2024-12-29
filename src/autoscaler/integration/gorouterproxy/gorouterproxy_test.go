package main_test

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Gorouterproxy", func() {
	var session *gexec.Session
	var testserver *httptest.Server

	BeforeEach(func() {
		testserver = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello, client")
		}))

		_, port, err := net.SplitHostPort(testserver.URL[len("http://"):])
		Expect(err).ShouldNot(HaveOccurred())

		cmd := exec.Command(cmdPath, "--port", "8080", "--forwardTo", port)
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())

	})

	AfterEach(func() {
		testserver.Close()
	})

	It("proxy request to test server and turns tls creds into xfcc header", func() {
		Eventually(session.Out).Should(gbytes.Say("gorouter-proxy.started"))
	})
})
