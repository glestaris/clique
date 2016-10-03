package dispatcher

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ice-stuff/clique/api"
	"github.com/ice-stuff/clique/scheduler"
	"github.com/ice-stuff/clique/transfer"
)

type TransferTask struct {
	Server       Interruptible
	Transferrer  Transferrer
	TransferSpec transfer.TransferSpec

	Registry ApiRegistry

	DesiredPriority int

	Logger *logrus.Logger

	done          bool
	transferState api.TransferState

	lock sync.Mutex
}

func (t *TransferTask) Run() {
	t.Server.Interrupt()
	defer t.Server.Resume()

	t.lock.Lock()
	t.transferState = api.TransferStateRunning
	t.lock.Unlock()

	res, err := t.Transferrer.Transfer(t.TransferSpec)
	if err != nil {
		t.Logger.Errorf("Transfer task will be rescheduled: %s", err.Error())

		t.lock.Lock()
		t.transferState = api.TransferStatePending
		t.lock.Unlock()

		return
	}

	t.Registry.RegisterResults(
		t.TransferSpec.IP,
		api.TransferResults{
			IP:        t.TransferSpec.IP,
			BytesSent: res.BytesSent,
			Checksum:  res.Checksum,
			Duration:  res.Duration,
			Time:      time.Now(),
		},
	)

	t.lock.Lock()
	t.transferState = api.TransferStateCompleted
	t.done = true
	t.lock.Unlock()
}

func (t *TransferTask) Priority() int {
	return t.DesiredPriority
}

func (t *TransferTask) State() scheduler.TaskState {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.done {
		return scheduler.TaskStateDone
	}

	return scheduler.TaskStateReady
}

func (t *TransferTask) TransferState() api.TransferState {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.transferState == "" {
		t.transferState = api.TransferStatePending
	}

	return t.transferState
}
