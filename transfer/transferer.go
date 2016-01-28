package transfer

import (
	"crypto/rand"
	"errors"
	"fmt"
	"hash/crc32"
	"net"
	"time"

	"github.com/Sirupsen/logrus"
)

var (
	ErrServerIsBusy = errors.New("server is busy")
)

type Transferer struct {
	Logger *logrus.Logger
}

func (c *Transferer) Transfer(spec TransferSpec) (TransferResults, error) {
	conn, err := net.Dial(
		"tcp", fmt.Sprintf("%s:%d", spec.IP.String(), spec.Port),
	)
	if err != nil {
		return TransferResults{}, fmt.Errorf("connecting to the server: %s", err)
	}
	defer conn.Close()
	c.Logger.Infof("Starting transfer to %s", conn.RemoteAddr().String())

	if err := c.handshake(conn); err != nil {
		return TransferResults{}, err
	}

	res, err := c.sendData(conn, spec.Size, c.randomBlock(1024))
	if err != nil {
		return TransferResults{}, err
	}
	c.Logger.WithFields(logrus.Fields{
		"duration":   res.Duration,
		"checksum":   res.Checksum,
		"bytes_sent": res.BytesSent,
	}).Info("Outgoing transfer is completed")

	return res, nil
}

func (c *Transferer) handshake(conn net.Conn) error {
	msgBytes := make([]byte, 16)
	n, _ := conn.Read(msgBytes)

	msg := string(msgBytes[:n])

	if msg == "ok" {
		return nil
	} else if msg == "i-am-busy" {
		return ErrServerIsBusy
	} else {
		return fmt.Errorf("unrecognized server response `%s`", msg)
	}
}

func (c *Transferer) randomBlock(size uint16) []byte {
	data := make([]byte, size)

	rand.Read(data)

	return data
}

func (c *Transferer) sendData(
	conn net.Conn,
	size uint32,
	block []byte,
) (TransferResults, error) {
	var (
		crc        uint32 = 0
		packetsAmt uint32 = size / 1024
		bytesSent  uint32 = 0
	)

	startTime := time.Now()
	for i := uint32(0); i < packetsAmt; i++ {
		n, _ := conn.Write(block)
		bytesSent += uint32(n)

		crc = crc32.Update(crc, crc32.IEEETable, block)
	}
	endTime := time.Now()

	return TransferResults{
		Duration:  endTime.Sub(startTime),
		BytesSent: bytesSent,
		Checksum:  crc,
	}, nil
}
