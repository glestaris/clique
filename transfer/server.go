package transfer

import (
	"fmt"
	"hash/crc32"
	"net"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

type server struct {
	listener net.Listener

	inProgress bool
	paused     bool

	lastTransfer TransferResults

	cond     *sync.Cond
	condLock *sync.Mutex
	lock     *sync.Mutex

	logger *logrus.Logger
}

func NewServer(logger *logrus.Logger, port uint16) (Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("listening to port %d: %s", port, err)
	}
	logger.Infof("Listening to port %d", port)

	condLock := new(sync.Mutex)
	return &server{
		listener: listener,

		inProgress: false,
		paused:     false,

		lastTransfer: TransferResults{},

		cond:     sync.NewCond(condLock),
		condLock: condLock,
		lock:     new(sync.Mutex),

		logger: logger,
	}, nil
}

func (s *server) Serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			break
		}

		s.lock.Lock()
		inProgress := s.inProgress
		paused := s.paused
		if !inProgress && !paused {
			s.inProgress = true
		}
		s.lock.Unlock()

		if paused || inProgress {
			go s.handleBusy(conn)
		} else {
			go s.handleTransfer(conn)
		}
	}
}

func (s *server) Pause() {
	s.lock.Lock()
	s.paused = true
	if !s.inProgress {
		s.lock.Unlock()
		return
	}

	s.condLock.Lock()
	s.lock.Unlock()
	defer s.condLock.Unlock()
	s.cond.Wait()

	return
}

func (s *server) Resume() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.paused = false

	return
}

func (s *server) Close() error {
	if s.listener == nil {
		return fmt.Errorf("server is not running")
	}

	return s.listener.Close()
}

func (s *server) LastTrasfer() TransferResults {
	s.condLock.Lock()
	defer s.condLock.Unlock()
	s.cond.Wait()

	s.lock.Lock()
	defer s.lock.Unlock()
	return s.lastTransfer
}

func (s *server) handleBusy(conn net.Conn) {
	defer conn.Close()

	conn.Write([]byte("i-am-busy"))
}

func (s *server) handleTransfer(conn net.Conn) {
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
	s.cond.Broadcast()
}

func (s *server) readData(conn net.Conn) TransferResults {
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
