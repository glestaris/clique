package acceptance_test

import (
	"fmt"

	"github.com/ice-stuff/clique/acceptance/runner"
	"github.com/ice-stuff/clique/config"
	"github.com/ice-stuff/clique/testhelpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

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
