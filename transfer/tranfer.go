package transfer

import (
	"net"
	"time"
)

//go:generate counterfeiter . Server
type Server interface {
	Serve()
	Pause()
	Resume()
	Close() error
	LastTrasfer() TransferResults
}

//go:generate counterfeiter . Transferer
type Transferer interface {
	Transfer(spec TransferSpec) (TransferResults, error)
}

type TransferSpec struct {
	IP   net.IP
	Port uint16
	Size uint32
}

type TransferResults struct {
	Duration  time.Duration
	Checksum  uint32
	BytesSent uint32
}
