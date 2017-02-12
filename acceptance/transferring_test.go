package acceptance_test

import (
	"net"
	"time"

	"github.com/ice-stuff/clique/acceptance/runner"
	"github.com/ice-stuff/clique/api"
	"github.com/ice-stuff/clique/config"
	"github.com/ice-stuff/clique/testhelpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transferring", func() {
	var (
		booTPort, booAPort, fooTPort uint16
		booClique, fooClique         *runner.ClqProcess
		booClient                    *api.Client
	)

	BeforeEach(func() {
		var err error

		booTPort = testhelpers.SelectPort(GinkgoParallelNode())
		booAPort = testhelpers.SelectPort(GinkgoParallelNode())
		booClique, err = startClique(config.Config{
			TransferPort: booTPort,
			APIPort:      booAPort,
		})
		Expect(err).NotTo(HaveOccurred())

		booClient = api.NewClient(
			"127.0.0.1", booAPort, time.Millisecond*100,
		)

		fooTPort = testhelpers.SelectPort(GinkgoParallelNode())
		fooClique, err = startClique(config.Config{
			TransferPort: fooTPort,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(booClique.Stop()).To(Succeed())
		Expect(fooClique.Stop()).To(Succeed())
	})

	It("should transfer the requested amount of bytes", func() {
		spec := api.TransferSpec{
			IP:   net.ParseIP("127.0.0.1"),
			Port: fooTPort,
			Size: 10 * 1024 * 1024,
		}
		Expect(booClient.CreateTransfer(spec)).To(Succeed())

		var resList []api.TransferResults
		Eventually(func() []api.TransferResults {
			var err error
			resList, err = booClient.TransferResultsByIP(net.ParseIP("127.0.0.1"))
			Expect(err).NotTo(HaveOccurred())
			return resList
		}, 5.0).Should(HaveLen(1))

		res := resList[0]
		Expect(res.IP).To(Equal(net.ParseIP("127.0.0.1")))
		// Iperf reported amount of bytes sent/received can differ than
		// requested. A 20% error is used here to verify that despite being
		// different, these numbers are not completely wrong.
		margin := float32(spec.Size) * 0.20
		Expect(res.BytesSent).To(BeNumerically("~", spec.Size, margin))
	})

	It("should populate the duration", func() {
		spec := api.TransferSpec{
			IP:   net.ParseIP("127.0.0.1"),
			Port: fooTPort,
			Size: 10 * 1024 * 1024,
		}
		Expect(booClient.CreateTransfer(spec)).To(Succeed())

		var resList []api.TransferResults
		Eventually(func() []api.TransferResults {
			var err error
			resList, err = booClient.TransferResultsByIP(net.ParseIP("127.0.0.1"))
			Expect(err).NotTo(HaveOccurred())
			return resList
		}, 5.0).Should(HaveLen(1))

		res := resList[0]
		Expect(res.IP).To(Equal(net.ParseIP("127.0.0.1")))
		Expect(res.Duration).NotTo(BeZero())
	})

	Context("when there are three clique agents", func() {
		var (
			mooAPort, mooTPort uint16
			mooClique          *runner.ClqProcess
			mooClient          *api.Client
		)

		BeforeEach(func() {
			var err error

			mooAPort = testhelpers.SelectPort(GinkgoParallelNode())
			mooTPort = testhelpers.SelectPort(GinkgoParallelNode())
			mooClique, err = startClique(config.Config{
				APIPort:      mooAPort,
				TransferPort: mooTPort,
			})
			Expect(err).NotTo(HaveOccurred())

			mooClient = api.NewClient(
				"127.0.0.1", mooAPort, time.Millisecond*100,
			)
		})

		AfterEach(func() {
			Expect(mooClique.Stop()).To(Succeed())
		})

		booPendingTransfers := func() []api.Transfer {
			transfers, err := booClient.TransfersByState(api.TransferStatePending)
			Expect(err).NotTo(HaveOccurred())
			return transfers
		}

		mooPendingTransfers := func() []api.Transfer {
			transfers, err := mooClient.TransfersByState(api.TransferStatePending)
			Expect(err).NotTo(HaveOccurred())
			return transfers
		}

		booRunningTransfers := func() []api.Transfer {
			transfers, err := booClient.TransfersByState(api.TransferStateRunning)
			Expect(err).NotTo(HaveOccurred())
			return transfers
		}

		mooRunningTransfers := func() []api.Transfer {
			transfers, err := mooClient.TransfersByState(api.TransferStateRunning)
			Expect(err).NotTo(HaveOccurred())
			return transfers
		}

		It("should serialize requests for outgoing transfers", func() {
			// boo sends to foo and moo at the same time
			Expect(booClient.CreateTransfer(api.TransferSpec{
				IP:   net.ParseIP("127.0.0.1"),
				Port: fooTPort,
				Size: 10 * 1024 * 1024,
			})).To(Succeed())
			Expect(booClient.CreateTransfer(api.TransferSpec{
				IP:   net.ParseIP("127.0.0.1"),
				Port: mooTPort,
				Size: 10 * 1024 * 1024,
			})).To(Succeed())

			Eventually(booPendingTransfers).Should(HaveLen(2))

			Eventually(booPendingTransfers, 2.0).Should(HaveLen(1))
			Eventually(booRunningTransfers, 2.0).Should(HaveLen(1))

			Eventually(booPendingTransfers, 5.0).Should(HaveLen(0))
		})

		It("should pend requests for incoming transfers when busy with outgoing", func() {
			// boo sends to foo and while the transfer is running, moo tries to send
			// to boo
			Expect(booClient.CreateTransfer(api.TransferSpec{
				IP:   net.ParseIP("127.0.0.1"),
				Port: fooTPort,
				Size: 100 * 1024 * 1024,
			})).To(Succeed())
			Eventually(booRunningTransfers).Should(HaveLen(1))

			Expect(mooClient.CreateTransfer(api.TransferSpec{
				IP:   net.ParseIP("127.0.0.1"),
				Port: booTPort,
				Size: 10 * 1024 * 1024,
			})).To(Succeed())
			Eventually(mooPendingTransfers).Should(HaveLen(1))

			Eventually(booPendingTransfers, 5.0).Should(HaveLen(0))
			Eventually(mooPendingTransfers, 5.0).Should(HaveLen(0))
		})

		It("should pend requests for outgoing transfers when busy with incoming", func() {
			// moo sends to boo and while the transfer is running, boo tries to send
			// to foo
			Expect(mooClient.CreateTransfer(api.TransferSpec{
				IP:   net.ParseIP("127.0.0.1"),
				Port: booTPort,
				Size: 100 * 1024 * 1024,
			})).To(Succeed())
			Eventually(mooRunningTransfers, 2.0).Should(HaveLen(1))

			Expect(booClient.CreateTransfer(api.TransferSpec{
				IP:   net.ParseIP("127.0.0.1"),
				Port: fooTPort,
				Size: 10 * 1024 * 1024,
			})).To(Succeed())
			Eventually(booPendingTransfers).Should(HaveLen(1))

			Eventually(booPendingTransfers, 5.0).Should(HaveLen(0))
			Eventually(mooPendingTransfers, 5.0).Should(HaveLen(0))
		})
	})
})
