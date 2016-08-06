package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/glestaris/clique"
	"github.com/labstack/echo"
)

//go:generate counterfeiter . Registry
type Registry interface {
	TransfersByState(state TransferState) []Transfer
	TransferResults() []TransferResults
	TransferResultsByIP(net.IP) []TransferResults
}

//go:generate counterfeiter . TransferCreator
type TransferCreator interface {
	Create(TransferSpec)
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

type Server struct {
	addr string

	listener   net.Listener
	httpServer *http.Server

	registry        Registry
	transferCreator TransferCreator

	lock sync.Mutex
}

func NewServer(
	port uint16,
	registry Registry,
	transferCreator TransferCreator,
) *Server {
	addr := fmt.Sprintf(":%d", port)

	s := &Server{
		addr: addr,

		registry:        registry,
		transferCreator: transferCreator,
	}

	e := echo.New()
	e.Get("/ping", s.handleGetPing)
	e.Get("/version", s.handleGetVersion)
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

func (s *Server) handleGetVersion(c *echo.Context) error {
	return c.String(200, clique.CliqueAgentVersion)
}

func (s *Server) handleGetTransfers(c *echo.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	state := ParseTransferState(c.Param("state"))

	res := s.registry.TransfersByState(state)

	return c.JSON(200, res)
}

func (s *Server) handleGetTransferResults(c *echo.Context) error {
	res := s.registry.TransferResults()

	return c.JSON(200, res)
}

func (s *Server) handleGetTransferResultsByIP(c *echo.Context) error {
	ip := net.ParseIP(c.Param("IP"))

	res := s.registry.TransferResultsByIP(ip)

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

	s.transferCreator.Create(spec)

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
