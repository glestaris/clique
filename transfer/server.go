package transfer

import (
	"fmt"
	"hash/crc32"
	"net"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

type Server struct {
	listener net.Listener

	inProgress bool
	paused     bool

	lastTransfer TransferResults

	transferFinish     *sync.Cond
	transferFinishLock *sync.Mutex
	lock               *sync.Mutex

	logger *logrus.Logger
}

func NewServer(logger *logrus.Logger, port uint16) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("listening to port %d: %s", port, err)
	}
	logger.Infof("Listening to port %d", port)

	transferFinishLock := new(sync.Mutex)
	return &Server{
		listener: listener,

		inProgress: false,
		paused:     false,

		lastTransfer: TransferResults{},

		transferFinish:     sync.NewCond(transferFinishLock),
		transferFinishLock: transferFinishLock,
		lock:               new(sync.Mutex),

		logger: logger,
	}, nil
}

func (s *Server) Serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.logger.Debugf(
				"Received (maybe expected) error while listening for connection: '%s'",
				err,
			)
			break
		}

		s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.paused || s.inProgress {
		go s.handleBusy(conn)
		return
	}

	s.inProgress = true
	go s.handleTransfer(conn)
}

func (s *Server) Interrupt() {
	s.lock.Lock()
	s.paused = true
	inProgress := s.inProgress
	s.lock.Unlock()

	if inProgress {
		s.transferFinishLock.Lock()
		defer s.transferFinishLock.Unlock()
		s.transferFinish.Wait()
	}
}

func (s *Server) Resume() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.paused = false

	return
}

func (s *Server) Close() error {
	if s.listener == nil {
		return fmt.Errorf("server is not running")
	}

	s.logger.Debug("Closing server...")
	return s.listener.Close()
}

func (s *Server) LastTrasfer() TransferResults {
	s.transferFinishLock.Lock()
	defer s.transferFinishLock.Unlock()
	s.transferFinish.Wait()

	s.lock.Lock()
	defer s.lock.Unlock()
	return s.lastTransfer
}

func (s *Server) handleBusy(conn net.Conn) {
	defer conn.Close()
	s.logger.Debugf("Server is busy for request from %s", conn.RemoteAddr().String())

	conn.Write([]byte("i-am-busy"))
}

func (s *Server) handleTransfer(conn net.Conn) {
	defer conn.Close()
	s.logger.Infof("Handling a transfer from %s", conn.RemoteAddr().String())

	conn.Write([]byte("ok"))

	res := s.readData(conn)
	s.logger.WithFields(logrus.Fields{
		"duration":   res.Duration,
		"checksum":   res.Checksum,
		"bytes_sent": res.BytesSent,
	}).Info("Incoming transfer is completed")

	s.lock.Lock()
	s.lastTransfer = res
	s.inProgress = false
	s.lock.Unlock()
	s.transferFinish.Broadcast()
}

func (s *Server) readData(conn net.Conn) TransferResults {
	defer conn.Close()

	var (
		res    = TransferResults{}
		buffer = make([]byte, 1024)
	)

	startTime := time.Now()
	for {
		n, err := conn.Read(buffer)
		if err != nil { // done reading
			break
		}
		res.BytesSent += uint32(n)

		res.Checksum = crc32.Update(res.Checksum, crc32.IEEETable, buffer)
	}
	endTime := time.Now()

	res.Duration = endTime.Sub(startTime)

	return res
}
