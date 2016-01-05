package dispatcher

import (
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/glestaris/ice-clique/api"
	"github.com/glestaris/ice-clique/scheduler"
	"github.com/glestaris/ice-clique/transfer"
)

const TransferTaskPriority int = 5

//go:generate counterfeiter . Scheduler
type Scheduler interface {
	Schedule(task scheduler.Task)
}

//go:generate counterfeiter . Interruptible
type Interruptible interface {
	Interrupt()
	Resume()
}

//go:generate counterfeiter . Transferer
type Transferer interface {
	Transfer(spec transfer.TransferSpec) (transfer.TransferResults, error)
}

//go:generate counterfeiter . ApiRegistry
type ApiRegistry interface {
	Register(ip net.IP, res api.TransferResults)
}

type Dispatcher struct {
	Scheduler Scheduler

	TransferServer Interruptible
	Transferer     Transferer

	ApiRegistry ApiRegistry

	Logger *logrus.Logger
}

func (d *Dispatcher) Create(spec api.TransferSpec) api.TransferStater {
	task := &TransferTask{
		Server:     d.TransferServer,
		Transferer: d.Transferer,
		TransferSpec: transfer.TransferSpec{
			IP:   spec.IP,
			Port: spec.Port,
			Size: spec.Size,
		},

		Registry: d.ApiRegistry,

		DesiredPriority: TransferTaskPriority,

		Logger: d.Logger,
	}

	d.Scheduler.Schedule(task)

	return task
}
