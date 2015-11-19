package transfer

import (
	"crypto/rand"
	"errors"
	"fmt"
	"hash/crc32"
	"net"
	"time"
)

type client struct{}

func NewClient() Transferer {
	return &client{}
}

func (c *client) Transfer(spec TransferSpec) (TransferResults, error) {
	conn, err := net.Dial(
		"tcp", fmt.Sprintf("%s:%d", spec.IP.String(), spec.Port),
	)
	if err != nil {
		return TransferResults{}, fmt.Errorf("connecting to the server: %s", err)
	}
	defer conn.Close()

	if err := c.handshake(conn); err != nil {
		return TransferResults{}, err
	}

	res, err := c.sendData(conn, spec.Size, c.randomBlock(1024))
	if err != nil {
		return TransferResults{}, err
	}

	return res, nil
}

func (c *client) handshake(conn net.Conn) error {
	msgBytes := make([]byte, 16)
	n, _ := conn.Read(msgBytes)

	msg := string(msgBytes[:n])

	if msg == "ok" {
		return nil
	} else if msg == "i-am-busy" {
		return errors.New("server is busy")
	} else {
		return fmt.Errorf("unrecognized server response `%s`", msg)
	}
}

func (c *client) randomBlock(size uint16) []byte {
	data := make([]byte, size)

	rand.Read(data)

	return data
}

func (c *client) sendData(
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
