package dispatcher_test

import (
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/glestaris/ice-clique/api"
	"github.com/glestaris/ice-clique/dispatcher"
	"github.com/glestaris/ice-clique/dispatcher/fakes"
	"github.com/glestaris/ice-clique/transfer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dispatcher", func() {
	var (
		fakeScheduler      *fakes.FakeScheduler
		fakeTransferServer *fakes.FakeInterruptible
		fakeTransferer     *fakes.FakeTransferer
		fakeApiRegistry    *fakes.FakeApiRegistry
		logger             *logrus.Logger
		dsptchr            *dispatcher.Dispatcher
	)

	BeforeEach(func() {
		fakeScheduler = new(fakes.FakeScheduler)
		logger = &logrus.Logger{
			Out:       GinkgoWriter,
			Level:     logrus.DebugLevel,
			Formatter: new(logrus.TextFormatter),
		}

		dsptchr = &dispatcher.Dispatcher{
			Scheduler: fakeScheduler,

			TransferServer: fakeTransferServer,
			Transferer:     fakeTransferer,

			ApiRegistry: fakeApiRegistry,

			Logger: nil,
		}
	})

	Describe("Create", func() {
		var (
			spec api.TransferSpec
		)

		BeforeEach(func() {
			spec = api.TransferSpec{
				IP:   net.ParseIP("127.88.91.234"),
				Port: 1212,
				Size: 10 * 1024 * 1024,
			}
		})

		It("should schedule a transfer task", func() {
			dsptchr.Create(spec)

			Expect(fakeScheduler.ScheduleCallCount()).To(Equal(1))
			task := fakeScheduler.ScheduleArgsForCall(0)

			Expect(task).To(BeAssignableToTypeOf(&dispatcher.TransferTask{}))
		})

		Describe("scheduled task", func() {
			var scheduledTask *dispatcher.TransferTask

			BeforeEach(func() {
				dsptchr.Create(spec)

				Expect(fakeScheduler.ScheduleCallCount()).To(Equal(1))
				task := fakeScheduler.ScheduleArgsForCall(0)

				var ok bool
				scheduledTask, ok = task.(*dispatcher.TransferTask)
				Expect(ok).To(BeTrue())
			})

			It("should be wired to the correct server and transferer", func() {
				Expect(scheduledTask.Server).To(Equal(fakeTransferServer))
				Expect(scheduledTask.Transferer).To(Equal(fakeTransferer))
			})

			It("should contain the correct tranfer spec", func() {
				Expect(scheduledTask.TransferSpec).To(Equal(transfer.TransferSpec{
					IP:   spec.IP,
					Port: spec.Port,
					Size: spec.Size,
				}))
			})

			It("should be wired to the correct registry", func() {
				Expect(scheduledTask.Registry).To(Equal(fakeApiRegistry))
			})

			It("should use the defined propery", func() {
				Expect(scheduledTask.DesiredPriority).To(Equal(dispatcher.TransferTaskPriority))
			})
		})
	})
})
