package dispatcher_test

import (
	"errors"
	"net"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/glestaris/ice-clique/dispatcher"
	"github.com/glestaris/ice-clique/dispatcher/fakes"
	"github.com/glestaris/ice-clique/scheduler"
	"github.com/glestaris/ice-clique/transfer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransferTask", func() {
	var (
		t              *dispatcher.TransferTask
		fakeServer     *fakes.FakeInterruptible
		fakeTransferer *fakes.FakeTransferer
		transferSpec   transfer.TransferSpec
		fakeRegistry   *fakes.FakeApiRegistry
		priority       int
		logger         *logrus.Logger
	)

	BeforeEach(func() {
		fakeServer = new(fakes.FakeInterruptible)
		fakeTransferer = new(fakes.FakeTransferer)
		transferSpec = transfer.TransferSpec{
			IP:   net.ParseIP("92.168.12.19"),
			Port: 1245,
			Size: 10 * 1024 * 1024,
		}
		fakeRegistry = new(fakes.FakeApiRegistry)
		priority = 10
		logger = &logrus.Logger{
			Out:       GinkgoWriter,
			Level:     logrus.DebugLevel,
			Formatter: new(logrus.TextFormatter),
		}

		t = &dispatcher.TransferTask{
			Server:          fakeServer,
			Transferer:      fakeTransferer,
			TransferSpec:    transferSpec,
			Registry:        fakeRegistry,
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

	It("should register the transfer results to the registry", func() {
		transferResults := transfer.TransferResults{
			Duration:  time.Millisecond * 100,
			Checksum:  uint32(12),
			BytesSent: uint32(10 * 1024 * 1024),
		}
		fakeTransferer.TransferReturns(transferResults, nil)

		t.Run()

		Expect(fakeRegistry.RegisterCallCount()).To(Equal(1))
		ip, res := fakeRegistry.RegisterArgsForCall(0)
		Expect(ip).To(Equal(transferSpec.IP))
		Expect(res.IP).To(Equal(transferSpec.IP))
		Expect(res.BytesSent).To(Equal(transferResults.BytesSent))
		Expect(res.Checksum).To(Equal(transferResults.Checksum))
		Expect(res.Duration).To(Equal(transferResults.Duration))
		Expect(res.Time).To(BeTemporally("~", time.Now(), time.Second))
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
