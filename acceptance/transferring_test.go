package acceptance_test

import (
	"fmt"
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/ice-stuff/clique/acceptance/runner"
	"github.com/ice-stuff/clique/config"
	"github.com/ice-stuff/clique/testhelpers"
	"github.com/ice-stuff/clique/transfer"
	"github.com/ice-stuff/clique/transfer/simple"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Single transferring", func() {
	var (
		transferClient *transfer.Client
		proc           *runner.ClqProcess
		port           uint16
	)

	BeforeEach(func() {
		var err error

		logger := &logrus.Logger{
			Out:       GinkgoWriter,
			Formatter: new(logrus.TextFormatter),
			Level:     logrus.InfoLevel,
		}
		transferConnector := transfer.NewConnector()
		transferSender := simple.NewSender(logger)
		transferClient = transfer.NewClient(
			logger, transferConnector, transferSender,
		)

		port = testhelpers.SelectPort(GinkgoParallelNode())

		proc, err = startClique(config.Config{
			TransferPort: port,
			RemoteHosts:  []string{},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(proc.Stop()).To(Succeed())
	})

	It("should accept transfers", func() {
		res, err := transferClient.Transfer(transfer.TransferSpec{
			IP:   net.ParseIP("127.0.0.1"),
			Port: port,
			Size: 10 * 1024 * 1024,
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(res.BytesSent).To(BeNumerically("==", 10*1024*1024))
	})

	It("should populate the duration", func() {
		res, err := transferClient.Transfer(transfer.TransferSpec{
			IP:   net.ParseIP("127.0.0.1"),
			Port: port,
			Size: 10 * 1024 * 1024,
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(res.Duration).NotTo(BeZero())
	})

	Context("when trying to run two transfers to the same server", func() {
		var transferStateCh chan string

		BeforeEach(func() {
			transferStateCh = make(chan string)

			go func(c chan string) {
				defer GinkgoRecover()

				c <- "started"

				Eventually(func() error {
					_, err := transferClient.Transfer(transfer.TransferSpec{
						IP:   net.ParseIP("127.0.0.1"),
						Port: port,
						Size: 10 * 1024 * 1024,
					})

					return err
				}).Should(Succeed())

				close(c)
			}(transferStateCh)
		})

		It("should reject the second one", func() {
			Eventually(transferStateCh).Should(Receive())

			Eventually(func() error {
				_, err := transferClient.Transfer(transfer.TransferSpec{
					IP:   net.ParseIP("127.0.0.1"),
					Port: port,
					Size: 10 * 1024,
				})

				return err
			}).Should(HaveOccurred())

			Eventually(transferStateCh, "15s").Should(BeClosed())
		})
	})
})

var _ = Describe("Logging", func() {
	var (
		tPortA, tPortB uint16
		procA, procB   *runner.ClqProcess
		hosts          []string
	)

	BeforeEach(func() {
		var err error

		tPortA = testhelpers.SelectPort(GinkgoParallelNode())
		tPortB = testhelpers.SelectPort(GinkgoParallelNode())
		hosts = []string{
			fmt.Sprintf("127.0.0.1:%d", tPortA),
			fmt.Sprintf("127.0.0.1:%d", tPortB),
		}

		procA, err = startClique(config.Config{
			TransferPort:     tPortA,
			RemoteHosts:      []string{hosts[1]},
			InitTransferSize: 1 * 1024 * 1024,
		}, "-debug")
		Expect(err).NotTo(HaveOccurred())

		procB, err = startClique(config.Config{
			TransferPort:     tPortB,
			RemoteHosts:      []string{hosts[0]},
			InitTransferSize: 1 * 1024 * 1024,
		}, "-debug")
		Expect(err).NotTo(HaveOccurred())
	})

	It("should print the clique result", func() {
		// wait to finish
		Eventually(procA.Buffer, 2.0).Should(gbytes.Say("new_state=done"))
		Eventually(procB.Buffer, 2.0).Should(gbytes.Say("new_state=done"))

		// exit to make sure all logs get pushed
		Expect(procA.Stop()).To(Succeed())
		Expect(procB.Stop()).To(Succeed())

		// grab contents
		procACont := string(procA.Buffer.Contents())
		procBCont := string(procB.Buffer.Contents())

		Expect(procACont).To(ContainSubstring("Incoming transfer is completed"))
		Expect(procACont).To(ContainSubstring("Outgoing transfer is completed"))

		Expect(procBCont).To(ContainSubstring("Incoming transfer is completed"))
		Expect(procBCont).To(ContainSubstring("Outgoing transfer is completed"))
	})
})
