package transfer

import (
	"net"
	"time"
)

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
