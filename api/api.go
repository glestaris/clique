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
	Time      time.Time     `json:"time"`
}

type TransferSpec struct {
	IP   net.IP `json:"ip"`
	Port uint16 `json:"port"`
	Size uint32 `json:"size"`
}
