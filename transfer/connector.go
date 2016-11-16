package transfer

import (
	"fmt"
	"net"
)

type connector struct {
}

func NewConnector() Connector {
	return &connector{}
}

func (c *connector) Connect(ip net.IP, port uint16) (net.Conn, error) {
	address := fmt.Sprintf("%s:%d", ip.String(), port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
