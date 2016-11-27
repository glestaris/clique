package iperf_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestIperf(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Iperf Suite")
}
