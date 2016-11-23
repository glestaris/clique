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

var _ = Describe("Api", func() {
	var (
		srcTPort, srcAPort uint16
		srcClique          *runner.ClqProcess
		srcClient          *api.Client
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
	})

	AfterEach(func() {
		Expect(srcClique.Stop()).To(Succeed())
	})

	Describe("Ping", func() {
		It("should succeed", func() {
			Expect(srcClient.Ping()).To(Succeed())
		})
	})

	Context("when there is a second clique-agent running", func() {
		var (
			destTPort  uint16
			destClique *runner.ClqProcess
		)

		BeforeEach(func() {
			var err error
			destTPort = testhelpers.SelectPort(GinkgoParallelNode())
			destClique, err = startClique(config.Config{
				TransferPort: destTPort,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(destClique.Stop()).To(Succeed())
		})

		It("should return list of pending transfers in the transfer source", func() {
			Expect(srcClient.CreateTransfer(api.TransferSpec{
				IP:   net.ParseIP("127.0.0.1"),
				Port: destTPort,
				Size: 10 * 1024 * 1024,
			})).To(Succeed())

			transferState := func() []api.Transfer {
				transfers, err := srcClient.TransfersByState(api.TransferStateRunning)
				Expect(err).NotTo(HaveOccurred())
				return transfers
			}

			// momentarily we see one running task
			Eventually(transferState).Should(HaveLen(1))

			// and then it finishes
			Eventually(transferState).Should(HaveLen(0))
		})
	})
})
