package simple_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSimple(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Simple Transfer Protocol Suite")
}
