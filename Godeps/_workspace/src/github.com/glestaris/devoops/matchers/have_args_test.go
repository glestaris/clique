package matchers_test

import (
	"github.com/glestaris/devoops"
	. "github.com/glestaris/devoops/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HaveArgs", func() {
	It("should match the arguments list", func() {
		argsList := []string{"i", "love", "bananas"}
		proc := devoops.Process{
			Args: argsList,
		}

		Expect(proc).To(HaveArgs(argsList...))
	})

	It("should not match a wrong arguments list", func() {
		proc := devoops.Process{
			Args: []string{"bananas", "are", "the", "best"},
		}

		Expect(proc).NotTo(HaveArgs("bananas", "the", "best"))
	})

	Context("when the actual is not a process object", func() {
		It("should error", func() {
			success, err := HaveArgs("banana").Match("banana")
			Expect(err).To(HaveOccurred())
			Expect(success).To(BeFalse())
		})
	})
})
