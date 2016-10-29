package acceptance_test

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ice-stuff/clique/acceptance/runner"
	"github.com/ice-stuff/clique/config"
	"github.com/ice-stuff/clique/transfer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Single transferring", func() {
	var (
		transferrer *transfer.Transferrer
		proc        *runner.ClqProcess
		port        uint16
	)

	BeforeEach(func() {
		var err error

		transferrer = &transfer.Transferrer{
			Logger: &logrus.Logger{
				Out:       GinkgoWriter,
				Formatter: new(logrus.TextFormatter),
				Level:     logrus.InfoLevel,
			},
		}

		port = uint16(5000 + rand.Intn(101) + GinkgoParallelNode())

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
		_, err := transferrer.Transfer(transfer.TransferSpec{
			IP:   net.ParseIP("127.0.0.1"),
			Port: port,
			Size: 10 * 1024 * 1024,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when trying to run two transfers to the same server", func() {
		var transferStateCh chan string

		BeforeEach(func() {
			transferStateCh = make(chan string)

			go func(c chan string) {
				defer GinkgoRecover()

				c <- "started"

				_, err := transferrer.Transfer(transfer.TransferSpec{
					IP:   net.ParseIP("127.0.0.1"),
					Port: port,
					Size: 20 * 1024 * 1024,
				})
				Expect(err).NotTo(HaveOccurred())

				close(c)
			}(transferStateCh)
		})

		It("should reject the second one", func() {
			Eventually(transferStateCh).Should(Receive())
			time.Sleep(time.Millisecond * 200) // synchronizing the requests

			_, err := transferrer.Transfer(transfer.TransferSpec{
				IP:   net.ParseIP("127.0.0.1"),
				Port: port,
				Size: 10 * 1024 * 1024,
			})
			Expect(err).To(HaveOccurred())

			Eventually(transferStateCh, "15s").Should(BeClosed())
		})
	})
})

var _ = Describe("Clique", func() {
	var (
		tPortA, tPortB uint16
		procA, procB   *runner.ClqProcess
		hosts          []string
	)

	BeforeEach(func() {
		var err error

		tPortA = uint16(5000 + rand.Intn(101) + GinkgoParallelNode())
		tPortB = uint16(5100 + rand.Intn(101) + GinkgoParallelNode())
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

	AfterEach(func() {
		Expect(procA.Stop()).To(Succeed())
		Expect(procB.Stop()).To(Succeed())
	})

	It("should print the clique result", func() {
		Eventually(procA.Buffer, 2.0).Should(gbytes.Say("new_state=done"))
		Eventually(procB.Buffer).Should(gbytes.Say("new_state=done"))

		procACont := procA.Buffer.Contents()
		Expect(procACont).To(ContainSubstring("Incoming transfer is completed"))
		Expect(procACont).To(ContainSubstring("Outgoing transfer is completed"))
		Expect(procACont).To(ContainSubstring(
			fmt.Sprintf("127.0.0.1:%d", tPortB),
		))

		procBCont := procB.Buffer.Contents()
		Expect(procBCont).To(ContainSubstring("Incoming transfer is completed"))
		Expect(procBCont).To(ContainSubstring("Outgoing transfer is completed"))
		Expect(procBCont).To(ContainSubstring(
			fmt.Sprintf("127.0.0.1:%d", tPortA),
		))
	})
})
