package acceptance_test

import (
	"net"
	"time"

	"github.com/glestaris/ice-clique/acceptance/runner"
	"github.com/glestaris/ice-clique/api"
	"github.com/glestaris/ice-clique/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Api", func() {
	var (
		tPort, aPort uint16
		clique       *runner.ClqProcess
		client       *api.Client
	)

	BeforeEach(func() {
		var err error

		tPort = uint16(5000 + GinkgoParallelNode())
		aPort = uint16(6000 + GinkgoParallelNode())

		clique, err = startClique(config.Config{
			TransferPort: tPort,
			APIPort:      aPort,
		})
		Expect(err).NotTo(HaveOccurred())

		client = api.NewClient(
			"127.0.0.1", aPort, time.Millisecond*100,
		)
	})

	AfterEach(func() {
		Expect(clique.Stop()).To(Succeed())
	})

	Describe("Ping", func() {
		It("should succeed", func() {
			Expect(client.Ping()).To(Succeed())
		})
	})

	Context("when there is a second clique-agent running", func() {
		var (
			tPortSecond, aPortSecond uint16
			cliqueSecond             *runner.ClqProcess
			clientSecond             *api.Client
		)

		BeforeEach(func() {
			var err error

			tPortSecond = uint16(5100 + GinkgoParallelNode())
			aPortSecond = uint16(6100 + GinkgoParallelNode())

			cliqueSecond, err = startClique(config.Config{
				TransferPort: tPortSecond,
				APIPort:      aPortSecond,
			})
			Expect(err).NotTo(HaveOccurred())

			clientSecond = api.NewClient(
				"127.0.0.1", aPortSecond, time.Millisecond*100,
			)
		})

		AfterEach(func() {
			Expect(cliqueSecond.Stop()).To(Succeed())
		})

		It("should transfer to the second clique-agent", func(done Done) {
			Expect(client.CreateTransfer(api.TransferSpec{
				IP:   net.ParseIP("127.0.0.1"),
				Port: tPortSecond,
				Size: 10 * 1024 * 1024,
			})).To(Succeed())

			var res api.TransferResults
			for {
				resList, err := client.TransferResultsByIP(net.ParseIP("127.0.0.1"))
				Expect(err).NotTo(HaveOccurred())

				if len(resList) != 0 {
					res = resList[0]
					break
				}
				time.Sleep(time.Millisecond * 100)
			}

			Expect(res.IP).To(Equal(net.ParseIP("127.0.0.1")))
			Expect(res.BytesSent).To(BeNumerically("==", 10*1024*1024))

			close(done)
		}, 5.0)
	})
})
