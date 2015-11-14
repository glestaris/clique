package acceptance_test

import (
	"io/ioutil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Acceptance", func() {
	Describe("Configuration", func() {
		var (
			configContents string
			configPath     string
		)

		BeforeEach(func() {
			configContents = ""
		})

		JustBeforeEach(func() {
			f, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			_, err = f.Write([]byte(configContents))
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())

			configPath = f.Name()
		})

		AfterEach(func() {
			Expect(os.Remove(configPath)).To(Succeed())
		})

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

			Context("and the configuration file is malformed", func() {
				BeforeEach(func() {
					configContents = "{'"
				})

				It("should fail", func() {
					session, err := gexec.Start(
						exec.Command(cliqueAgentBin, "-config", configPath),
						GinkgoWriter, GinkgoWriter,
					)
					Expect(err).NotTo(HaveOccurred())

					Consistently(session).ShouldNot(gexec.Exit(0))
				})
			})

			Context("and the configuration file is valid", func() {
				BeforeEach(func() {
					configContents = "{}"
				})

				It("should succeed", func() {
					session, err := gexec.Start(
						exec.Command(cliqueAgentBin, "-config", configPath),
						GinkgoWriter, GinkgoWriter,
					)
					Expect(err).NotTo(HaveOccurred())

					Eventually(session).Should(gexec.Exit(0))
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
})
