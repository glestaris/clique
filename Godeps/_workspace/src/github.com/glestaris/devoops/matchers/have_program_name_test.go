package matchers_test

import (
	"github.com/glestaris/devoops"
	. "github.com/glestaris/devoops/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HaveProgramName", func() {
	It("should match the program name", func() {
		proc := devoops.Process{
			ProgramName: "banana",
		}

		Expect(proc).To(HaveProgramName("banana"))
	})

	It("should not match a wrong program name", func() {
		proc := devoops.Process{
			ProgramName: "banana",
		}

		Expect(proc).NotTo(HaveProgramName("apple"))
	})

	Context("when the actual is not a process object", func() {
		It("should error", func() {
			success, err := HaveProgramName("banana").Match("banana")
			Expect(err).To(HaveOccurred())
			Expect(success).To(BeFalse())
		})
	})
})
