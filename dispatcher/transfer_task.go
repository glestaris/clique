package dispatcher

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/glestaris/ice-clique/api"
	"github.com/glestaris/ice-clique/scheduler"
	"github.com/glestaris/ice-clique/transfer"
)

type TransferTask struct {
	Server       Interruptible
	Transferer   Transferer
	TransferSpec transfer.TransferSpec

	Registry ApiRegistry

	DesiredPriority int

	Logger *logrus.Logger

	done bool
}

func (t *TransferTask) Run() {
	t.Server.Interrupt()
	defer t.Server.Resume()

	res, err := t.Transferer.Transfer(t.TransferSpec)
	if err != nil {
		t.Logger.Errorf("Transfer task will be rescheduled: %s", err.Error())
		return
	}

	t.Registry.Register(
		t.TransferSpec.IP,
		api.TransferResults{
			IP:        t.TransferSpec.IP,
			BytesSent: res.BytesSent,
			Checksum:  res.Checksum,
			Duration:  res.Duration,
			Time:      time.Now(),
		},
	)

	t.done = true
}

func (t *TransferTask) Priority() int {
	return t.DesiredPriority
}

func (t *TransferTask) State() scheduler.TaskState {
	if t.done {
		return scheduler.TaskStateDone
	}

	return scheduler.TaskStateReady
}
