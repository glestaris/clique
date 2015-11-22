package transfer

import (
	"fmt"
	"hash/crc32"
	"net"
	"sync"
	"time"
)

type server struct {
	listener net.Listener

	inProgress   bool
	lastTransfer TransferResults

	cond *sync.Cond
	lock *sync.Mutex
}

func NewServer(port uint16) (Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("listening to port %d: %s", port, err)
	}

	lock := new(sync.Mutex)
	return &server{
		listener: listener,
		lock:     lock,
		cond:     sync.NewCond(lock),
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
		if !inProgress {
			s.inProgress = true
		}
		s.lock.Unlock()

		if inProgress {
			go s.handleBusy(conn)
		} else {
			go s.handleTransfer(conn)
		}
	}
}

func (s *server) Close() error {
	if s.listener == nil {
		return fmt.Errorf("server is not running")
	}

	return s.listener.Close()
}

func (s *server) LastTrasfer() TransferResults {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.cond.Wait()
	return s.lastTransfer
}

func (s *server) handleBusy(conn net.Conn) {
	defer conn.Close()

	conn.Write([]byte("i-am-busy"))
}

func (s *server) handleTransfer(conn net.Conn) {
	defer conn.Close()

	conn.Write([]byte("ok"))

	res := s.readData(conn)

	s.lock.Lock()
	s.lastTransfer = res
	s.inProgress = false
	s.cond.Signal()
	s.lock.Unlock()
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
