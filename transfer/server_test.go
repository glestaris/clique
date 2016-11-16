package transfer_test

import (
	"errors"
	"io"
	"net"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ice-stuff/clique/transfer"
	"github.com/ice-stuff/clique/transfer/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {
	var (
		logger               *logrus.Logger
		fakeListener         *fakes.FakeListener
		fakeTransferReceiver *fakes.FakeTransferReceiver
		server               *transfer.Server

		listenerConnChan chan net.Conn
		listenerErrChan  chan error
		serverClosed     chan struct{}
	)

	BeforeEach(func() {
		logger = &logrus.Logger{
			Out:       GinkgoWriter,
			Level:     logrus.DebugLevel,
			Formatter: new(logrus.TextFormatter),
		}
		fakeListener = new(fakes.FakeListener)
		fakeTransferReceiver = new(fakes.FakeTransferReceiver)
		server = transfer.NewServer(logger, fakeListener, fakeTransferReceiver)

		listenerConnChan = make(chan net.Conn, 100)
		listenerErrChan = make(chan error)
		fakeListener.AcceptStub = func() (net.Conn, error) {
			select {
			case conn := <-listenerConnChan:
				return conn, nil
			case err := <-listenerErrChan:
				return nil, err
			}
		}
	})

	JustBeforeEach(func() {
		serverClosed = make(chan struct{})
		go func() {
			server.Serve()
			close(serverClosed)
		}()
	})

	AfterEach(func() {
		listenerErrChan <- errors.New("close now!")
		Eventually(serverClosed).Should(BeClosed())
		close(listenerConnChan)
		close(listenerErrChan)
	})

	Describe("Serve", func() {
		It("should accept a connection", func() {
			conn, _ := net.Pipe()
			listenerConnChan <- conn

			Eventually(fakeListener.AcceptCallCount).Should(Equal(2))
			// 2 = one received connection and one blocking call
		})

		It("should process the connection", func() {
			pushedConn, _ := net.Pipe()
			listenerConnChan <- pushedConn

			Eventually(fakeTransferReceiver.ReceiveTransferCallCount).Should(Equal(1))
			Expect(fakeTransferReceiver.ReceiveTransferArgsForCall(0)).To(
				Equal(pushedConn),
			)
		})
	})

	Describe("LastTransfer", func() {
		var resultsChan chan transfer.TransferResults

		BeforeEach(func() {
			resultsChan = make(chan transfer.TransferResults, 100)
			fakeTransferReceiver.ReceiveTransferStub = func(_ io.ReadWriter) (
				transfer.TransferResults, error,
			) {
				return <-resultsChan, nil
			}
		})

		AfterEach(func() {
			close(resultsChan)
		})

		It("blocks until a transfer is handled by the server", func() {
			conn, _ := net.Pipe()
			listenerConnChan <- conn

			receivedResultsChan := make(chan transfer.TransferResults, 100)
			go func() {
				receivedResultsChan <- server.LastTransfer()
			}()
			Consistently(receivedResultsChan).ShouldNot(Receive())

			pushedResults := transfer.TransferResults{
				Duration: time.Second,
			}
			resultsChan <- pushedResults
			Eventually(receivedResultsChan).Should(Receive(Equal(pushedResults)))
		})
	})
})
