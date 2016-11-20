package simple

import (
	"crypto/rand"
	"fmt"
	"hash/crc32"
	"io"
	"time"

	"github.com/Sirupsen/logrus"
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
	if err := s.handshake(conn); err != nil {
		return transfer.TransferResults{}, err
	}

	randomData, err := s.randomBlock(1024)
	if err != nil {
		return transfer.TransferResults{}, err
	}

	return s.sendData(conn, spec.Size, randomData)
}

func (s *Sender) handshake(conn io.ReadWriter) error {
	msgBytes := make([]byte, 16)
	n, err := conn.Read(msgBytes)
	if err != nil {
		return err
	}

	msg := string(msgBytes[:n])
	if msg == "ok" {
		return nil
	} else if msg == "i-am-busy" {
		return ErrBusy
	} else {
		return fmt.Errorf("unrecognized server response `%s`", msg)
	}
}

func (s *Sender) randomBlock(size uint16) ([]byte, error) {
	randomData := make([]byte, size)
	if _, err := rand.Read(randomData); err != nil {
		return nil, err
	}

	return randomData, nil
}

func (s *Sender) sendData(conn io.ReadWriter, size uint32, block []byte) (
	transfer.TransferResults, error,
) {
	res := transfer.TransferResults{}
	packetsAmt := uint32(size / 1024)

	startTime := time.Now()
	for i := uint32(0); i < packetsAmt; i++ {
		n, err := conn.Write(block)
		if err != nil {
			return transfer.TransferResults{}, err
		}
		res.BytesSent += uint32(n)

		res.Checksum = crc32.Update(res.Checksum, crc32.IEEETable, block)
	}
	endTime := time.Now()

	res.Duration = endTime.Sub(startTime)

	return res, nil
}
