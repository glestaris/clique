package simpletransfer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSimpletransfer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Simple Transfer Protocol Suite")
}
