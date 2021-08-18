package display_test

import (
	"changelog/display"
	"changelog/github"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckArray(t *testing.T) {
	assert.False(t, display.ArrayContains([]string{"one", "two"}, "three"))
	assert.True(t, display.ArrayContains([]string{"one", "two"}, "two"))
}

func TestGenerateOutput(t *testing.T) {
	prs := []github.PullRequest{
		{Number: 1, Title: "Test PR 1", Author: "me", Labels: []string{"breaking-change"}},
		{Number: 2, Title: "Test PR 2", Author: "me", Labels: []string{"enhancement"}},
	}

	expectedChangeLog := `# Changelog for app-autoscaler-release


## Breaking Changes

* [Test PR 1]() - **me**

## Enhancements

* [Test PR 2]() - **me**
`

	changelog, nextVersion, err := display.GenerateOutput(prs, "1.0.0")
	assert.NoError(t, err)
	assert.Equal(t, expectedChangeLog, changelog)
	assert.Equal(t, "2.0.0", nextVersion)
}
