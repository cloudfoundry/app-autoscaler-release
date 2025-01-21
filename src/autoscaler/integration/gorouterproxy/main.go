package main

import (
	"crypto/tls"
	"encoding/pem"
	"io"
	"log"
	"net/http"
	"os"

	"flag"
	"fmt"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers/auth"
)

var (
	port      = flag.String("port", "8888", "Port for xfcc proxy")
	forwardTo = flag.String("forwardTo", "", "Port to forward to")
	logger    = log.New(os.Stdout, "gorouter-proxy", log.LstdFlags)
)

func main() {
	flag.Parse()
	startServer()
}

func startServer() {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", *port),
		Handler: http.HandlerFunc(forwardHandler),
		TLSConfig: &tls.Config{
			ClientAuth: tls.RequireAnyClientCert,
		},
	}

	certFile, keyFile := getCertFiles()
	logger.Printf("starting gorouter-proxy on port %s, forwarding to %s", *port, *forwardTo)
	if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
		logger.Printf("Error starting server: %v", err)
	}
}

func getCertFiles() (string, string) {
	testCertDir := "../../../../test-certs"
	return testCertDir + "/gorouter.crt", testCertDir + "/gorouter.key"
}

func forwardHandler(w http.ResponseWriter, inRequest *http.Request) {
	tls := inRequest.TLS
	if !isClientCertValid(tls) {
		http.Error(w, "No client certificate", http.StatusForbidden)
		return
	}

	cert := createCert(tls)
	if cert == nil {
		http.Error(w, "Failed to parse client certificate", http.StatusInternalServerError)
		return
	}

	resp, err := forwardRequest(cert)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	if body, err := io.ReadAll(resp.Body); err == nil {
		w.Write(body)
	}
}

func isClientCertValid(tls *tls.ConnectionState) bool {
	if tls == nil || len(tls.PeerCertificates) == 0 {
		logger.Printf("No client certificate")
		return false
	}
	logger.Print("received tls: ", tls.PeerCertificates)
	return true
}

func createCert(tls *tls.ConnectionState) *auth.Cert {
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: tls.PeerCertificates[0].Raw,
	})
	return auth.NewCert(string(pemData))
}

func forwardRequest(cert *auth.Cert) (*http.Response, error) {
	client := &http.Client{}
	url := fmt.Sprintf("http://localhost:%s", *forwardTo)
	outRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	outRequest.Header.Add("X-Forwarded-Client-Cert", cert.GetXFCCHeader())
	return client.Do(outRequest)
}
