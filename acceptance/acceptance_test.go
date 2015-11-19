package acceptance_test

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/glestaris/ice-clique/acceptance/runner"
	"github.com/glestaris/ice-clique/config"
	"github.com/glestaris/ice-clique/transfer"
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

			Context("and options are missing", func() {
				BeforeEach(func() {
					configContents = "{}"
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
					configContents = fmt.Sprintf(`{
						"transfer_port": %d
					}`, 5000+GinkgoParallelNode())
				})

				It("should succeed", func() {
					buffer := gbytes.NewBuffer()
					session, err := gexec.Start(
						exec.Command(cliqueAgentBin, "-config", configPath),
						buffer, buffer,
					)
					Expect(err).NotTo(HaveOccurred())

					Eventually(buffer).Should(gbytes.Say("iCE Clique Agent"))

					session.Terminate()
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

	Describe("Single transferring", func() {
		var (
			cfg    config.Config
			client transfer.Transferer
			proc   *runner.ClqProcess
		)

		BeforeEach(func() {
			cfg = config.Config{
				TransferPort: 5000 + uint16(GinkgoParallelNode()),
			}

			client = transfer.NewClient()
		})

		JustBeforeEach(func() {
			var err error

			proc, err = startClique(cfg)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(proc.Stop()).To(Succeed())
		})

		It("should accept transfers", func() {
			_, err := client.Transfer(transfer.TransferSpec{
				IP:   net.ParseIP("127.0.0.1"),
				Port: cfg.TransferPort,
				Size: 10 * 1024 * 1024,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when trying to run two transfers to the same server", func() {
			var transferStateCh chan string

			JustBeforeEach(func() {
				transferStateCh = make(chan string)

				go func(c chan string) {
					defer GinkgoRecover()

					c <- "started"

					_, err := client.Transfer(transfer.TransferSpec{
						IP:   net.ParseIP("127.0.0.1"),
						Port: cfg.TransferPort,
						Size: 50 * 1024 * 1024,
					})
					Expect(err).NotTo(HaveOccurred())

					close(c)
				}(transferStateCh)
			})

			It("should reject the second one", func() {
				Eventually(transferStateCh).Should(Receive())
				time.Sleep(time.Millisecond * 200) // synchronizing the requests

				_, err := client.Transfer(transfer.TransferSpec{
					IP:   net.ParseIP("127.0.0.1"),
					Port: cfg.TransferPort,
					Size: 10 * 1024 * 1024,
				})
				Expect(err).To(HaveOccurred())

				Eventually(transferStateCh, "15s").Should(BeClosed())
			})
		})
	})
})
