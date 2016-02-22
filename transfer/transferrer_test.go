package transfer_test

import (
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/glestaris/ice-clique/transfer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transferrer", func() {
	var logger *logrus.Logger

	BeforeEach(func() {
		logger = &logrus.Logger{
			Out:       GinkgoWriter,
			Level:     logrus.DebugLevel,
			Formatter: new(logrus.TextFormatter),
		}
	})

	Describe("Transfer", func() {
		Context("when there is no server", func() {
			It("should return an error", func() {
				transferrer := &transfer.Transferrer{Logger: logger}
				spec := transfer.TransferSpec{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 12121,
				}
				_, err := transferrer.Transfer(spec)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
