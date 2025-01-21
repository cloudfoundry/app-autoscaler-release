package main_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers/auth"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/testhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Gorouterproxy", func() {
	var (
		session    *gexec.Session
		testserver *httptest.Server
		proxyPort  string
		orgGUID    string
		spaceGUID  string
	)

	BeforeEach(func() {
		orgGUID = "valid-org"
		spaceGUID = "valid-space"
		proxyPort = fmt.Sprintf("%d", 8888+GinkgoParallelProcess())
		testserver = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Forwarded-Client-Cert") == "" {
				http.Error(w, "No xfcc header", http.StatusForbidden)
				return
			}

			err := auth.CheckAuth(r, orgGUID, spaceGUID)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "Hello, client")
		}))

		_, port, err := net.SplitHostPort(testserver.URL[len("http://"):])
		Expect(err).ShouldNot(HaveOccurred())

		cmd := exec.Command(cmdPath, "--port", proxyPort, "--forwardTo", port)
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())

	})

	AfterEach(func() {
		session.Kill().Wait()
		testserver.Close()
	})

	It("proxy request to test server and turns tls creds into xfcc header", func() {
		Eventually(session.Out, 20*time.Second).Should(gbytes.Say("starting"))

		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)

		key := testhelpers.GenerateClientKeyWithPrivateKey(privateKey)

		cert, err := testhelpers.GenerateClientCertWithPrivateKey(orgGUID, spaceGUID, privateKey)
		Expect(err).ToNot(HaveOccurred())

		testCertDir := "../../../../test-certs"

		rootCertFile := testCertDir + "/gorouter-ca.crt"

		tlsCert, err := tls.X509KeyPair(cert, key)
		Expect(err).ToNot(HaveOccurred())

		caCert, err := ioutil.ReadFile(rootCertFile)
		if err != nil {
			panic(fmt.Sprintf("Failed to load CA certificate: %v", err))
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		c := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates:       []tls.Certificate{tlsCert},
					RootCAs:            caCertPool,
					InsecureSkipVerify: true,
				},
			},
		}

		Expect(proxyPort).ToNot(BeEmpty())
		resp, err := c.Get(fmt.Sprintf("https://127.0.0.1:%s", proxyPort))
		Expect(err).ShouldNot(HaveOccurred())

		body, err := io.ReadAll(resp.Body)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(string(body)).To(ContainSubstring("Hello, client"))
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
	})
})
