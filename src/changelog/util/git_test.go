package util_test

import (
	"changelog/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetShaOfSubmoduleAtCommit(t *testing.T) {
	util.Runner = &FakeCommandRunner{}
	sha, err := util.GetShaOfSubmoduleAtCommit("HEAD")
	assert.NoError(t, err)
	assert.Equal(t, sha, "e57e6dccf303e621ec4239b5d97aeba168a6de50")
}

type FakeCommandRunner struct {
}

// RunWithoutRetry Execute the command without retrying on failure and block waiting for return values.
func (d FakeCommandRunner) RunWithoutRetry(c *util.Command) (string, error) {
	return "160000 commit e57e6dccf303e621ec4239b5d97aeba168a6de50\t../app-autoscaler", nil
}
