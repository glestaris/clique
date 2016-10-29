package api

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/ice-stuff/clique"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/engine/standard"
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

	httpServer engine.Server

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

	s.httpServer = standard.New(addr)
	s.httpServer.SetHandler(e)

	return s
}

func (s *Server) handleGetPing(c echo.Context) error {
	return c.String(200, "")
}

func (s *Server) handleGetVersion(c echo.Context) error {
	return c.String(200, clique.CliqueAgentVersion)
}

func (s *Server) handleGetTransfers(c echo.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	state := ParseTransferState(c.Param("state"))

	res := s.registry.TransfersByState(state)

	return c.JSON(200, res)
}

func (s *Server) handleGetTransferResults(c echo.Context) error {
	res := s.registry.TransferResults()

	return c.JSON(200, res)
}

func (s *Server) handleGetTransferResultsByIP(c echo.Context) error {
	ip := net.ParseIP(c.Param("IP"))

	res := s.registry.TransferResultsByIP(ip)

	return c.JSON(200, res)
}

func (s *Server) handlePostTransfers(c echo.Context) error {
	req := c.Request()
	decoder := json.NewDecoder(req.Body())

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
	return s.httpServer.Start()
}

func (s *Server) Close() error {
	return s.httpServer.Stop()
}
