package transfer

import (
	"io"
	"net"

	"github.com/Sirupsen/logrus"
)

//go:generate counterfeiter . TransferSender
type TransferSender interface {
	SendTransfer(spec TransferSpec, conn io.ReadWriter) (TransferResults, error)
}

//go:generate counterfeiter . Connector
type Connector interface {
	Connect(ip net.IP, port uint16) (net.Conn, error)
}

type Client struct {
	logger         *logrus.Logger
	connector      Connector
	transferSender TransferSender
}

func NewClient(
	logger *logrus.Logger, connector Connector,
	transferSender TransferSender,
) *Client {
	return &Client{
		logger:         logger,
		connector:      connector,
		transferSender: transferSender,
	}
}

func (c *Client) Transfer(spec TransferSpec) (TransferResults, error) {
	conn, err := c.connector.Connect(spec.IP, spec.Port)
	if err != nil {
		c.logger.Errorf("Failed to connect to server: '%s'", err)
		return TransferResults{}, err
	}
	defer conn.Close()

	c.logger.Infof("Starting transfer to %s", conn.RemoteAddr().String())
	res, err := c.transferSender.SendTransfer(spec, conn)
	if err != nil {
		c.logger.Errorf("Failed to send transfer: '%s'", err)
		return TransferResults{}, err
	}

	c.logger.WithFields(logrus.Fields{
		"duration":   res.Duration,
		"checksum":   res.Checksum,
		"bytes_sent": res.BytesSent,
		"rtt":        res.RTT,
	}).Info("Outgoing transfer is completed")
	return res, nil
}
