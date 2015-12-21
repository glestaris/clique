package devoops_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDevoops(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Devoops Suite")
}
