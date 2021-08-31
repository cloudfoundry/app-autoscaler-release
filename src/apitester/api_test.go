package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestCurler_Curl(t *testing.T) {
	curler := Curler{
		NumAllowedErrors: 10,
		NumRequests: 3000,
		Timeout: 10*time.Second,
		//Gap: 2*time.Second,
		Url: "https://api.autoscaler.ci.cloudfoundry.org/v2/info",
		SkipSslValidation: true,
	}

	curler.Start()

	fmt.Printf("Errors %d/%d\n", curler.NumActualErrors, curler.NumAllowedErrors)

	assert.Equal(t, 0, curler.NumActualErrors, "Should not have received any errors")
}

func HelloServer(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is an example server.\n"))
	// fmt.Fprintf(w, "This is an example server.\n")
	// io.WriteString(w, "This is an example server.\n")
}
func TestCurler_LocalCurl(t *testing.T) {
	http.HandleFunc("/v2/info", HelloServer)
	fmt.Printf("before server")
	go func() {
		err := http.ListenAndServeTLS("localhost:8082", "test_data/server.crt", "test_data/server.key", nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
		fmt.Printf("Started server")
	}()

	curler := Curler{
		NumAllowedErrors: 10,
		NumRequests:      3000,
		Timeout:          10 * time.Second,
		//Gap: 2*time.Second,
		Url:               "https://localhost:8082/v2/info",
		SkipSslValidation: true,
	}

	curler.Start()

	fmt.Printf("Errors %d/%d\n", curler.NumActualErrors, curler.NumAllowedErrors)

	assert.Equal(t, 0, curler.NumActualErrors, "Should not have received any errors")
}
