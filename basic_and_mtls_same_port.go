package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// Define username and password for Basic Auth
const (
	basicAuthUsername = "admin"
	basicAuthPassword = "password"
)

// BasicAuthMiddleware handles Basic Authentication
func BasicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("basic")
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Basic" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		creds, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		pair := strings.SplitN(string(creds), ":", 2)
		if len(pair) != 2 || pair[0] != basicAuthUsername || pair[1] != basicAuthPassword {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// mTLSAuthMiddleware ensures mutual TLS authentication
func mTLSAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("mtls")
		if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// do the cert validation now...

		next.ServeHTTP(w, r)
	})
}

// combinedAuthMiddleware decides which authentication method to use
func combinedAuthMiddleware(nextMTLS http.Handler, nextBasicAuth http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(strings.ToLower(r.URL.String()), "health") {
			BasicAuthMiddleware(nextBasicAuth).ServeHTTP(w, r)

			return
		}

		mTLSAuthMiddleware(nextMTLS).ServeHTTP(w, r)
	})
}

func handlerMTLS(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world! (mtls protected)")
}

func handlerBasicAuth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world! (basic auth protected)")
}

func main() {
	// Load server certificate and key
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatalf("failed to load server certificate: %v", err)
	}

	// Load CA certificate for client certificate verification
	caCert, err := os.ReadFile("ca.crt")
	if err != nil {
		log.Fatalf("failed to read CA certificate: %v", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Configure TLS settings
	tlsConfig := &tls.Config{
		ClientCAs:    caCertPool,
		ClientAuth:   tls.VerifyClientCertIfGiven,
		Certificates: []tls.Certificate{cert},
	}
	tlsConfig.BuildNameToCertificate()

	server := &http.Server{
		Addr:      ":8443",
		Handler:   combinedAuthMiddleware(http.HandlerFunc(handlerMTLS), http.HandlerFunc(handlerBasicAuth)),
		TLSConfig: tlsConfig,
	}

	log.Println("Starting server on port 8443")
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
