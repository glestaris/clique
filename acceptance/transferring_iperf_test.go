package acceptance_test

import (
	"runtime"
	"time"

	"github.com/ice-stuff/clique/acceptance/runner"
	"github.com/ice-stuff/clique/api"
	"github.com/ice-stuff/clique/config"
	"github.com/ice-stuff/clique/testhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transferring with Iperf", func() {
	var (
		srcTPort, srcAPort, destTPort uint16
		srcClique, destClique         *runner.ClqProcess
		srcClient                     *api.Client
	)

	BeforeEach(func() {
		if !useIperf {
			Skip("This test can only run with Iperf.")
		}
		if runtime.GOOS != "linux" {
			Skip("This test can only run with Linux.")
		}

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
})
