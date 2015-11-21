package transfer_test

import (
	"errors"
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/glestaris/ice-clique/scheduler"
	"github.com/glestaris/ice-clique/transfer"
	"github.com/glestaris/ice-clique/transfer/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransferTask", func() {
	var (
		t              *transfer.TransferTask
		fakeServer     *fakes.FakeServer
		fakeTransferer *fakes.FakeTransferer
		transferSpec   transfer.TransferSpec
		priority       int
		logger         *logrus.Logger
	)

	BeforeEach(func() {
		fakeServer = new(fakes.FakeServer)
		fakeTransferer = new(fakes.FakeTransferer)
		transferSpec = transfer.TransferSpec{
			IP:   net.ParseIP("92.168.12.19"),
			Port: 1245,
			Size: 10 * 1024 * 1024,
		}
		priority = 10
		logger = &logrus.Logger{
			Out:       GinkgoWriter,
			Level:     logrus.DebugLevel,
			Formatter: new(logrus.TextFormatter),
		}

		t = &transfer.TransferTask{
			Server:          fakeServer,
			Transferer:      fakeTransferer,
			TransferSpec:    transferSpec,
			DesiredPriority: priority,
			Logger:          logger,
		}
	})

	It("should run the transfer", func() {
		t.Run()
		Expect(fakeTransferer.TransferCallCount()).To(Equal(1))
		Expect(fakeTransferer.TransferArgsForCall(0)).To(Equal(transferSpec))
	})

	It("should pause the server", func() {
		t.Run()
	})

	It("should resume the server", func() {
		t.Run()
		Expect(fakeServer.ResumeCallCount()).To(Equal(1))
	})

	It("should not change state as long as the transfer fails", func() {
		fakeTransferer.TransferReturns(
			transfer.TransferResults{}, errors.New("banana"),
		)

		for i := 0; i < 100; i++ {
			t.Run()
			Expect(t.State()).To(Equal(scheduler.TaskStateReady))
		}

		fakeTransferer.TransferReturns(transfer.TransferResults{}, nil)
		t.Run()
		Expect(t.State()).To(Equal(scheduler.TaskStateDone))
	})

	It("should return the provided priority", func() {
		Expect(t.Priority()).To(Equal(priority))
	})

	It("should change state appropriately", func() {
		Expect(t.State()).To(Equal(scheduler.TaskStateReady))
		t.Run()
		Expect(t.State()).To(Equal(scheduler.TaskStateDone))
	})
})
