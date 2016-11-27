package iperf

import (
	"fmt"
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/ice-stuff/clique/iperf/runner"
	"github.com/ice-stuff/clique/transfer"
)

type Sender struct {
	logger *logrus.Logger
}

func NewSender(logger *logrus.Logger) *Sender {
	return &Sender{
		logger: logger,
	}
}

func (s *Sender) SendTransfer(spec transfer.TransferSpec, conn io.ReadWriter) (
	transfer.TransferResults, error,
) {
	iperfPort, err := s.handshake(conn)
	if err != nil {
		return transfer.TransferResults{}, err
	}

	return runner.RunTest(runner.ClientConfig{
		// Transfer target
		TargetHostIP:   spec.IP,
		TargetHostPort: iperfPort,
		// Transfer size
		BufferSize: 1024,
		BytesAmt:   uint(spec.Size),
	})
}

func (s *Sender) handshake(conn io.ReadWriter) (uint16, error) {
	msgBytes := make([]byte, 16)
	n, err := conn.Read(msgBytes)
	if err != nil {
		return 0, err
	}

	msg := string(msgBytes[:n])
	var iperfPort uint16
	_, err = fmt.Sscanf(msg, "ok - %d", &iperfPort)
	if err != nil {
		if msg == "i-am-busy" {
			return 0, ErrBusy
		} else {
			return 0, fmt.Errorf("unrecognized server response `%s`: %s", msg, err)
		}
	}

	return iperfPort, nil
}
