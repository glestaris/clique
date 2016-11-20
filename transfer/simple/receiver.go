package simple

import (
	"hash/crc32"
	"io"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ice-stuff/clique/transfer"
)

type Receiver struct {
	logger *logrus.Logger

	stateMutex          sync.Mutex
	isBusy              bool
	isPaused            bool
	transferFinishMutex sync.Mutex
	transferFinish      sync.Cond
}

func NewReceiver(logger *logrus.Logger) *Receiver {
	return &Receiver{
		logger: logger,
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

func (r *Receiver) handleBusy(conn io.ReadWriter) error {
	r.logger.Debugf("Server is busy!")
	if _, err := conn.Write([]byte("i-am-busy")); err != nil {
		return err
	}

	return nil
}

func (r *Receiver) handleTransfer(conn io.ReadWriter) (
	transfer.TransferResults, error,
) {
	if _, err := conn.Write([]byte("ok")); err != nil {
		return transfer.TransferResults{}, err
	}

	res := transfer.TransferResults{}
	buffer := make([]byte, 1024)

	startTime := time.Now()
	for {
		n, err := conn.Read(buffer)
		if err != nil { // done reading
			break
		}

		res.BytesSent += uint32(n)
		res.Checksum = crc32.Update(res.Checksum, crc32.IEEETable, buffer)
	}
	endTime := time.Now()

	res.Duration = endTime.Sub(startTime)

	return res, nil
}
