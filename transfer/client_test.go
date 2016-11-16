package transfer_test

import (
	"errors"
	"net"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ice-stuff/clique/transfer"
	"github.com/ice-stuff/clique/transfer/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var (
		logger             *logrus.Logger
		fakeConnector      *fakes.FakeConnector
		conn               net.Conn
		fakeTransferSender *fakes.FakeTransferSender
		client             *transfer.Client
	)

	BeforeEach(func() {
		logger = &logrus.Logger{
			Out:       GinkgoWriter,
			Level:     logrus.DebugLevel,
			Formatter: new(logrus.TextFormatter),
		}

		fakeConnector = new(fakes.FakeConnector)
		conn, _ = net.Pipe()
		fakeConnector.ConnectReturns(conn, nil)

		fakeTransferSender = new(fakes.FakeTransferSender)

		client = transfer.NewClient(logger, fakeConnector, fakeTransferSender)
	})

	It("should create a connection to the server", func() {
		spec := transfer.TransferSpec{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 1200,
		}
		_, err := client.Transfer(spec)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeConnector.ConnectCallCount()).To(Equal(1))
		receivedIP, receivedPort := fakeConnector.ConnectArgsForCall(0)
		Expect(receivedIP).To(Equal(spec.IP))
		Expect(receivedPort).To(Equal(spec.Port))
	})

	It("should use the transfer sender to send a transfer", func() {
		spec := transfer.TransferSpec{
			Size: 10 * 1024,
		}
		_, err := client.Transfer(spec)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeTransferSender.SendTransferCallCount()).To(Equal(1))
		receivedSpec, _ := fakeTransferSender.SendTransferArgsForCall(0)
		Expect(receivedSpec).To(Equal(spec))
	})

	It("should use the connection provided by the connector", func() {
		_, err := client.Transfer(transfer.TransferSpec{})
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeTransferSender.SendTransferCallCount()).To(Equal(1))
		_, receivedConn := fakeTransferSender.SendTransferArgsForCall(0)
		Expect(receivedConn).To(Equal(conn))
	})

	It("should return the transfer results", func() {
		fakeTransferResults := transfer.TransferResults{
			Duration:  time.Second,
			Checksum:  12222,
			BytesSent: 28448,
		}
		fakeTransferSender.SendTransferReturns(fakeTransferResults, nil)

		receivedTransferResults, err := client.Transfer(transfer.TransferSpec{})
		Expect(err).NotTo(HaveOccurred())

		Expect(receivedTransferResults).To(Equal(fakeTransferResults))
	})

	Context("when it failes to enstablish a connection", func() {
		var connectErr error

		BeforeEach(func() {
			connectErr = errors.New("failed to enstablish connection")
			fakeConnector.ConnectReturns(nil, connectErr)
		})

		It("should return the error", func() {
			_, err := client.Transfer(transfer.TransferSpec{})
			Expect(err).To(Equal(connectErr))
		})
	})

	Context("when it fails to conduct the transfer", func() {
		var senderErr error

		BeforeEach(func() {
			senderErr = errors.New("failed to conduct the transfer")
			fakeTransferSender.SendTransferReturns(
				transfer.TransferResults{}, senderErr,
			)
		})

		It("should return the error", func() {
			_, err := client.Transfer(transfer.TransferSpec{})
			Expect(err).To(Equal(senderErr))
		})
	})
})
