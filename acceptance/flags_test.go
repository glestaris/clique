package acceptance_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Flags", func() {
	Context("when `-config` option is passed", func() {
		Context("and the configuration file does not exist", func() {
			It("should fail", func() {
				session, err := gexec.Start(
					exec.Command(cliqueAgentBin, "-config", "/path/to/banana/cfg"),
					GinkgoWriter, GinkgoWriter,
				)
				Expect(err).NotTo(HaveOccurred())

				Consistently(session).ShouldNot(gexec.Exit(0))
			})
		})
	})

	Context("when `-config` option is not passed", func() {
		It("should fail", func() {
			session, err := gexec.Start(
				exec.Command(cliqueAgentBin),
				GinkgoWriter, GinkgoWriter,
			)
			Expect(err).NotTo(HaveOccurred())

			Consistently(session).ShouldNot(gexec.Exit(0))
		})
	})
})
