package api

import (
	"net"
	"time"
)

type TransferResults struct {
	IP        net.IP        `json:"ip"`
	BytesSent uint32        `json:"bytes_sent"`
	Checksum  uint32        `json:"checksum"`
	Duration  time.Duration `json:"duration"`
	RTT       time.Duration `json:"rtt"`
	Time      time.Time     `json:"time"`
}

type TransferSpec struct {
	IP   net.IP `json:"ip"`
	Port uint16 `json:"port"`
	Size uint32 `json:"size"`
}

type TransferState string

func (state TransferState) String() string {
	return string(state)
}

const (
	TransferStatePending   TransferState = "pending"
	TransferStateRunning   TransferState = "running"
	TransferStateCompleted TransferState = "completed"
	TransferStateUnknown   TransferState = "unknown"
)

func ParseTransferState(state string) TransferState {
	switch state {
	case "pending":
		return TransferStatePending
	case "running":
		return TransferStateRunning
	case "completed":
		return TransferStateCompleted
	default:
		return TransferStateUnknown
	}
}

type Transfer struct {
	Spec  TransferSpec  `json:"spec"`
	State TransferState `json:"state"`
}
