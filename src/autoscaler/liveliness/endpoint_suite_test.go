package liveliness_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLiveliness(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Liveliness endpoint suite")
}