package runner

// #include "runner.h"
import "C"
import (
	"net"
	"time"
)

type Config struct {
	// The interval time in seconds between periodic bandwidth, jitter, and loss
	// measurements.
	MeasurementInterval time.Duration
}

func (c Config) ToIRConfig() C.IRConfig {
	return C.IRConfig{
		measurement_interval: C.int(c.MeasurementInterval.Seconds()),
	}
}

type ServerConfig struct {
	Config
	// Port to listen to
	ListenPort uint16
}

func (c ServerConfig) ToIRServerConfig() C.IRServerConfig {
	return C.IRServerConfig{
		ir_config:   c.Config.ToIRConfig(),
		listen_port: C.int(c.ListenPort),
	}
}

type Protocol int

const (
	ProtocolTCP = Protocol(iota)
	ProtocolUDP
	ProtocolSCTP
)

func (p Protocol) ToIRProtocol() C.IRProtocol {
	return C.IRProtocol(p)
}

type ClientConfig struct {
	Config
	// Target  to connect to
	TargetHostIP   net.IP
	TargetHostPort uint16
	// Transport protocol to use for the measurements.
	Protocol Protocol
	// Duration of the stream. Default: 10 seconds.
	Duration time.Duration
	// Amount of buffers to send.
	BytesAmt uint
	// Size of transmission buffer. Default: 128 KB for TCP and 8 KB for UDP.
	BufferSize uint
	// Amount of packets to send.
	PacketsAmt uint
}

func (c ClientConfig) ToIRClientConfig() C.IRClientConfig {
	return C.IRClientConfig{
		ir_config:        c.Config.ToIRConfig(),
		target_host_ip:   C.CString(c.TargetHostIP.String()),
		target_host_port: C.int(c.TargetHostPort),
		protocol:         c.Protocol.ToIRProtocol(),
		duration_secs:    C.int(c.Duration.Seconds()),
		bytes_amt:        C.int(c.BytesAmt),
		buffer_size:      C.int(c.BufferSize),
		packets_amt:      C.int(c.PacketsAmt),
	}
}
