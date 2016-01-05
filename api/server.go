package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/labstack/echo"
)

//go:generate counterfeiter . TransferResultsRegistry
type TransferResultsRegistry interface {
	TransferResults() []TransferResults
	TransferResultsByIP(net.IP) []TransferResults
}

//go:generate counterfeiter . TransferStater
type TransferStater interface {
	TransferState() TransferState
}

//go:generate counterfeiter . TransferCreator
type TransferCreator interface {
	Create(TransferSpec) TransferStater
}

type SECode string

const (
	SERegistryFailed SECode = "registry-failed"
	SEInvalidRequst         = "invalid-request"
	SECreateFialed          = "create-failed"
)

type ServerError struct {
	Code SECode `json:"code"`
	Msg  string `json:"msg"`
}

type liveTransfer struct {
	spec       TransferSpec
	savedState TransferState
	stater     TransferStater
}

func (t *liveTransfer) state() TransferState {
	if t.savedState != TransferStateCompleted {
		t.savedState = t.stater.TransferState()
		if t.savedState == TransferStateCompleted {
			t.stater = nil
		}
	}

	return t.savedState
}

func (t *liveTransfer) transfer() Transfer {
	return Transfer{
		Spec:  t.spec,
		State: t.state(),
	}
}

type Server struct {
	addr string

	listener   net.Listener
	httpServer *http.Server

	transferResultsRegistry TransferResultsRegistry
	transferCreator         TransferCreator

	liveTransfers []liveTransfer

	lock sync.Mutex
}

func NewServer(
	port uint16,
	transferResultsRegistry TransferResultsRegistry,
	transferCreator TransferCreator,
) *Server {
	addr := fmt.Sprintf(":%d", port)

	s := &Server{
		addr: addr,

		transferResultsRegistry: transferResultsRegistry,
		transferCreator:         transferCreator,
	}

	e := echo.New()
	e.Get("/ping", s.handleGetPing)
	e.Get("/transfers/:state", s.handleGetTransfers)
	e.Get("/transfer_results", s.handleGetTransferResults)
	e.Get("/transfer_results/:IP", s.handleGetTransferResultsByIP)
	e.Post("/transfers", s.handlePostTransfers)
	s.httpServer = e.Server(addr)
	s.httpServer.SetKeepAlivesEnabled(false)

	return s
}

func (s *Server) handleGetPing(c *echo.Context) error {
	return c.String(200, "")
}

func (s *Server) handleGetTransfers(c *echo.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	state := ParseTransferState(c.Param("state"))

	res := []Transfer{}
	for _, lt := range s.liveTransfers {
		if lt.state() == state {
			res = append(res, lt.transfer())
		}
	}

	return c.JSON(200, res)
}

func (s *Server) handleGetTransferResults(c *echo.Context) error {
	res := s.transferResultsRegistry.TransferResults()

	return c.JSON(200, res)
}

func (s *Server) handleGetTransferResultsByIP(c *echo.Context) error {
	ip := net.ParseIP(c.Param("IP"))

	res := s.transferResultsRegistry.TransferResultsByIP(ip)

	return c.JSON(200, res)
}

func (s *Server) handlePostTransfers(c *echo.Context) error {
	req := c.Request()
	decoder := json.NewDecoder(req.Body)

	var spec TransferSpec
	if err := decoder.Decode(&spec); err != nil {
		// untested return
		return c.JSON(
			400, &ServerError{
				Code: SEInvalidRequst,
				Msg:  fmt.Sprintf("Invalid transfer spec: %s", err),
			},
		)
	}

	transferStater := s.transferCreator.Create(spec)
	s.lock.Lock()
	s.liveTransfers = append(s.liveTransfers, liveTransfer{
		spec:   spec,
		stater: transferStater,
	})
	s.lock.Unlock()

	return c.String(200, "")
}

func (s *Server) Serve() error {
	var err error

	s.lock.Lock()
	s.listener, err = net.Listen("tcp", s.addr)
	s.lock.Unlock()

	if err != nil {
		return fmt.Errorf("listening to address '%s': %s", s.addr, err)
	}

	s.httpServer.Serve(s.listener)

	return nil
}

func (s *Server) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.listener == nil {
		return errors.New("server is closed already")
	}

	s.listener.Close()
	s.listener = nil

	return nil
}
