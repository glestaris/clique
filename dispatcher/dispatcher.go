package dispatcher

import (
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/glestaris/ice-clique/api"
	"github.com/glestaris/ice-clique/api/registry"
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

//go:generate counterfeiter . Transferrer
type Transferrer interface {
	Transfer(spec transfer.TransferSpec) (transfer.TransferResults, error)
}

//go:generate counterfeiter . ApiRegistry
type ApiRegistry interface {
	RegisterTransfer(spec api.TransferSpec, stater registry.TransferStater)
	RegisterResults(ip net.IP, res api.TransferResults)
}

type Dispatcher struct {
	Scheduler Scheduler

	TransferServer Interruptible
	Transferrer     Transferrer

	ApiRegistry ApiRegistry

	Logger *logrus.Logger
}

func (d *Dispatcher) Create(spec api.TransferSpec) {
	task := &TransferTask{
		Server:     d.TransferServer,
		Transferrer: d.Transferrer,
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
	d.ApiRegistry.RegisterTransfer(spec, task)
}
