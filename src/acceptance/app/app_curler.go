package app

import (
	"context"
	"fmt"
	"github.com/onsi/gomega"
	"os/exec"
	"time"
)

type Curler struct {
	NumAllowedErrors int
	NumActualErrors int
	UriCreator uriCreator
	CurlConfig CurlConfig
}

func NewAppCurler(cfg CurlConfig) Curler {
	uriCreator := &AppUriCreator{CurlConfig: cfg}
	return Curler{
		CurlConfig: cfg,
		NumAllowedErrors: 10,
		NumActualErrors: 0,
		UriCreator: uriCreator,
	}
}

func (a *Curler) Curl(appName string, path string, timeout time.Duration, args ...string) string {
	appUri := a.UriCreator.AppUri(appName, path)
	curlArgs := append([]string{appUri}, args...)
	curlArgs = append([]string{"-s"}, curlArgs...)
	curlArgs = append([]string{"-H", "Expect:"}, curlArgs...)

	if a.CurlConfig.GetSkipSSLValidation() {
		curlArgs = append([]string{"-k"}, curlArgs...)
	}

	// Create a new context and add a timeout to it
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel() // The cancel should be deferred so resources are cleaned up

	// Create the command with our context
	cmd := exec.CommandContext(ctx, "curl", curlArgs...)

	// This time we can simply use Output() to get the result.
	out, err := cmd.Output()

	// We want to check the context error to see if the timeout was executed.
	// The error returned by cmd.Output() will be OS specific based on what
	// happens when a process is killed.
	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Command timed out")
		a.NumActualErrors++
		return ""
	}

	// If there's no context error, we know the command completed (or errored).
	fmt.Println("Output:", string(out))
	if err != nil {
		fmt.Println("Non-zero exit code:", err)
		a.NumActualErrors++
	}

	gomega.Expect(a.NumActualErrors).ShouldNot(gomega.Equal(a.NumAllowedErrors))

	return string(out)
}