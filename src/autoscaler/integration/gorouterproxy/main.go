package main

import (
	"net/http"

	"flag"
	"fmt"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers/auth"
)

// redirects any request to the HTTP version of the URL on a different port, all in localhost.

var port = flag.String("port", "8888", "Port for xfcc proxy")
var forwardTo = flag.String("forwardTo", "", "Port to forward to")

func main() {
	flag.Parse()

	fmt.Printf("gorouter-proxy.started - port: %s - forwardTo: %s", *port, *forwardTo)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", *port),
		Handler: http.HandlerFunc(forwardHandler),
	}

	server.ListenAndServe()
}

func forwardHandler(w http.ResponseWriter, inRequest *http.Request) {
	tls := inRequest.TLS
	cert := auth.NewCert(string(tls.PeerCertificates[0].Raw))

	client := &http.Client{}
	outRequest, err := http.NewRequest("GET", *forwardTo, nil)

	outRequest.Header.Add("X-Forwarded-Client-Cert", cert.GetXFCCHeader())
	client.Do(outRequest)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
