package main

import (
	"crypto/tls"
	"encoding/pem"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"flag"
	"fmt"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers/auth"
)

var (
	port      = flag.String("port", "8888", "Port for xfcc proxy")
	forwardTo = flag.String("forwardTo", "", "Port to forward to")
	keyFile   = flag.String("keyFile", "", "Path to key file")
	certFile  = flag.String("certFile", "", "Path to cert file")
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
			MinVersion: tls.VersionTLS12,
		},
		ReadHeaderTimeout: 10 * time.Second,
	}

	certFile, keyFile := getCertFiles()

	logger.Printf("gorouter-proxy.started - port %s, forwardTo %s", *port, *forwardTo)
	if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
		logger.Printf("Error starting server: %v", err)
	}
}

func getCertFiles() (string, string) {
	if *keyFile != "" && *certFile != "" {
		return *certFile, *keyFile
	}

	testCertDir := "../../../../test-certs"
	return testCertDir + "/gorouter.crt", testCertDir + "/gorouter.key"
}

func forwardHandler(w http.ResponseWriter, inRequest *http.Request) {
	var body []byte
	var err error

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

	if body, err = io.ReadAll(resp.Body); err != nil {
		logger.Printf("Error reading response: %v", err)
		return
	}

	if _, err := w.Write(body); err != nil {
		logger.Printf("Error writing response: %v", err)
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
	logger.Printf("Forwarding request to %s", *forwardTo)
	url := fmt.Sprintf("http://127.0.0.1:%s", *forwardTo)
	outRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	outRequest.Header.Add("X-Forwarded-Client-Cert", cert.GetXFCCHeader())
	return client.Do(outRequest)
}
