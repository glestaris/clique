package dispatcher

import (
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/ice-stuff/clique/api"
	"github.com/ice-stuff/clique/api/registry"
	"github.com/ice-stuff/clique/scheduler"
	"github.com/ice-stuff/clique/transfer"
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

//go:generate counterfeiter . TransferClient
type TransferClient interface {
	Transfer(spec transfer.TransferSpec) (transfer.TransferResults, error)
}

//go:generate counterfeiter . ApiRegistry
type ApiRegistry interface {
	RegisterTransfer(spec api.TransferSpec, stater registry.TransferStater)
	RegisterResults(ip net.IP, res api.TransferResults)
}

type Dispatcher struct {
	Scheduler Scheduler

	TransferInterruptible Interruptible
	TransferClient        TransferClient

	ApiRegistry ApiRegistry

	Logger *logrus.Logger
}

func (d *Dispatcher) Create(spec api.TransferSpec) {
	d.Logger.WithFields(logrus.Fields{
		"ip":   spec.IP,
		"port": spec.Port,
		"size": spec.Size,
	}).Debug("Received new task")

	task := &TransferTask{
		TransferInterruptible: d.TransferInterruptible,
		TransferClient:        d.TransferClient,
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
