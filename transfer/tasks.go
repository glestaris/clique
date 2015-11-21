package transfer

import (
	"github.com/Sirupsen/logrus"
	"github.com/glestaris/ice-clique/scheduler"
)

type TransferTask struct {
	Server          Server
	Transferer      Transferer
	TransferSpec    TransferSpec
	DesiredPriority int
	Logger          *logrus.Logger
	done            bool
}

func (t *TransferTask) Run() {
	t.Server.Pause()
	defer t.Server.Resume()

	if _, err := t.Transferer.Transfer(t.TransferSpec); err != nil {
		t.Logger.Errorf("Transfer task will be rescheduled: %s", err.Error())
		return
	}

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
