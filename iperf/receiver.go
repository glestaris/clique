package iperf

import (
	"fmt"
	"io"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/ice-stuff/clique/iperf/runner"
	"github.com/ice-stuff/clique/transfer"
)

type Receiver struct {
	logger    *logrus.Logger
	iperfPort uint16

	isBusy   bool
	isPaused bool

	stateMutex          *sync.Mutex
	transferFinishMutex *sync.Mutex
	transferFinish      *sync.Cond
}

func NewReceiver(logger *logrus.Logger, iperfPort uint16) *Receiver {
	transferFinishMutex := new(sync.Mutex)
	return &Receiver{
		logger:    logger,
		iperfPort: iperfPort,

		isBusy:   false,
		isPaused: false,

		stateMutex:          new(sync.Mutex),
		transferFinishMutex: transferFinishMutex,
		transferFinish:      sync.NewCond(transferFinishMutex),
	}
}

func (r *Receiver) ReceiveTransfer(conn io.ReadWriter) (
	transfer.TransferResults, error,
) {
	r.stateMutex.Lock()
	isBusy := r.isPaused || r.isBusy
	if !isBusy {
		r.isBusy = true
	}
	r.stateMutex.Unlock()

	if isBusy {
		if err := r.handleBusy(conn); err != nil {
			r.logger.Errorf("Failed to send busy message: %s", err)
		}

		return transfer.TransferResults{}, ErrBusy
	}

	defer func() {
		// reset state
		r.stateMutex.Lock()
		r.isBusy = false
		r.stateMutex.Unlock()
		r.transferFinish.Broadcast()
	}()
	return r.handleTransfer(conn)
}

func (r *Receiver) Interrupt() {
	r.stateMutex.Lock()
	r.isPaused = true
	isBusy := r.isBusy
	r.stateMutex.Unlock()

	if isBusy {
		r.transferFinishMutex.Lock()
		defer r.transferFinishMutex.Unlock()
		r.transferFinish.Wait()
	}
}

func (r *Receiver) Resume() {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()
	r.isPaused = false
}

func (r *Receiver) IsBusy() bool {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()

	return r.isBusy || r.isPaused
}

func (r *Receiver) handleBusy(conn io.ReadWriter) error {
	r.logger.Debug("[IPERF] Server is busy!")
	if _, err := conn.Write([]byte("i-am-busy")); err != nil {
		return err
	}

	return nil
}

func (r *Receiver) handleTransfer(conn io.ReadWriter) (
	transfer.TransferResults, error,
) {
	r.logger.Debug("[IPERF] Handling the transfer...")
	msg := fmt.Sprintf("ok - %d", r.iperfPort)
	if _, err := conn.Write([]byte(msg)); err != nil {
		return transfer.TransferResults{}, err
	}

	return runner.ListenAndServe(runner.ServerConfig{
		ListenPort: r.iperfPort,
	})
}
