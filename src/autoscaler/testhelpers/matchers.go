package testhelpers

import (
	"fmt"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

type ErrorMessage struct{ types.GomegaMatcher }

func (matcher *ErrorMessage) Match(actual interface{}) (success bool, err error) {
	// is purely nil?
	if actual == nil {
		return false, nil
	}

	// must be an 'error' type
	err, ok := actual.(error)
	if !ok {
		return false, fmt.Errorf("Expected an error-type.  Got:\n%s", format.Object(actual, 1))
	}
	if err == nil {
		return false, fmt.Errorf("Expected an error.  Got:\nnil")
	}
	return matcher.GomegaMatcher.Match(err.Error())
}

func (matcher *ErrorMessage) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected an error to have occurred.  Got:\n%s", format.Object(actual, 1))
}

func (matcher *ErrorMessage) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Unexpected error:\n%s\n%s\n%s", format.Object(actual, 1), format.IndentString(actual.(error).Error(), 1), "occurred")
}

func HaveErrorMessage(matcher types.GomegaMatcher) types.GomegaMatcher {
	return &ErrorMessage{matcher}
}
