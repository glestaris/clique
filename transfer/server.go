package transfer

import (
	"io"
	"net"

	"github.com/Sirupsen/logrus"
)

//go:generate counterfeiter . TransferReceiver
type TransferReceiver interface {
	ReceiveTransfer(conn io.ReadWriter) (TransferResults, error)
}

type Server struct {
	logger           *logrus.Logger
	listener         net.Listener
	transferReceiver TransferReceiver

	resChan chan TransferResults
}

func NewServer(
	logger *logrus.Logger, listener net.Listener,
	transferReceiver TransferReceiver,
) *Server {
	return &Server{
		logger:           logger,
		listener:         listener,
		transferReceiver: transferReceiver,

		resChan: make(chan TransferResults, 1024),
	}
}

func (s *Server) Serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.logger.Debugf(
				"Received (maybe expected) error while listening for connection: '%s'",
				err,
			)
			return
		}

		go func() {
			s.logger.Infof("Handling a transfer from %s", conn.RemoteAddr().String())
			res, err := s.transferReceiver.ReceiveTransfer(conn)
			if err != nil {
				conn.Close()
				s.logger.Debugf("Failed to process connection: '%s'", err)
				return
			}
			conn.Close()

			s.logger.WithFields(logrus.Fields{
				"duration":   res.Duration,
				"checksum":   res.Checksum,
				"bytes_sent": res.BytesSent,
			}).Info("Incoming transfer is completed")
			s.resChan <- res
		}()
	}
}

func (s *Server) LastTransfer() TransferResults {
	return <-s.resChan
}
