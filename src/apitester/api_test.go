package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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
