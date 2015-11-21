package acceptance_test

import (
	"net"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/glestaris/ice-clique/acceptance/runner"
	"github.com/glestaris/ice-clique/config"
	"github.com/glestaris/ice-clique/transfer"
)

var _ = Describe("Acceptance", func() {
	Describe("Flags", func() {
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
