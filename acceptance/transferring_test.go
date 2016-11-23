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
		srcTPort, srcAPort, destTPort uint16
		srcClique, destClique         *runner.ClqProcess
		srcClient                     *api.Client
	)

	BeforeEach(func() {
		var err error

		srcTPort = testhelpers.SelectPort(GinkgoParallelNode())
		srcAPort = testhelpers.SelectPort(GinkgoParallelNode())
		srcClique, err = startClique(config.Config{
			TransferPort: srcTPort,
			APIPort:      srcAPort,
		})
		Expect(err).NotTo(HaveOccurred())

		srcClient = api.NewClient(
			"127.0.0.1", srcAPort, time.Millisecond*100,
		)

		destTPort = testhelpers.SelectPort(GinkgoParallelNode())
		destClique, err = startClique(config.Config{
			TransferPort: destTPort,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(srcClique.Stop()).To(Succeed())
		Expect(destClique.Stop()).To(Succeed())
	})

	It("should transfer the requested amount of bytes", func() {
		spec := api.TransferSpec{
			IP:   net.ParseIP("127.0.0.1"),
			Port: destTPort,
			Size: 10 * 1024 * 1024,
		}
		Expect(srcClient.CreateTransfer(spec)).To(Succeed())

		var resList []api.TransferResults
		Eventually(func() []api.TransferResults {
			var err error
			resList, err = srcClient.TransferResultsByIP(net.ParseIP("127.0.0.1"))
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
			Port: destTPort,
			Size: 10 * 1024 * 1024,
		}
		Expect(srcClient.CreateTransfer(spec)).To(Succeed())

		var resList []api.TransferResults
		Eventually(func() []api.TransferResults {
			var err error
			resList, err = srcClient.TransferResultsByIP(net.ParseIP("127.0.0.1"))
			Expect(err).NotTo(HaveOccurred())
			return resList
		}, 5.0).Should(HaveLen(1))

		res := resList[0]
		Expect(res.IP).To(Equal(net.ParseIP("127.0.0.1")))
		Expect(res.Duration).NotTo(BeZero())
	})
})
