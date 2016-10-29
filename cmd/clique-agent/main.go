package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"code.cloudfoundry.org/clock"
	"github.com/Sirupsen/logrus"
	"github.com/ice-stuff/clique"
	"github.com/ice-stuff/clique/api"
	"github.com/ice-stuff/clique/api/registry"
	"github.com/ice-stuff/clique/config"
	"github.com/ice-stuff/clique/dispatcher"
	"github.com/ice-stuff/clique/scheduler"
	"github.com/ice-stuff/clique/transfer"
)

var (
	cfg config.Config

	configPath = flag.String("config", "", "The configuration file path")

	version = flag.Bool("version", false, "Print clique-agent version")
	debug   = flag.Bool("debug", false, "Print debug messaging")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("clique-agent v%s\n", clique.CliqueAgentVersion)
		os.Exit(0)
	}

	level := logrus.InfoLevel
	if *debug {
		level = logrus.DebugLevel
	}
	logger := &logrus.Logger{
		Out:       os.Stdout,
		Level:     level,
		Formatter: new(logrus.TextFormatter),
	}

	if *configPath == "" {
		logger.Fatal("`-config` option is required")
	}
	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		logger.Fatal(err.Error())
	}

	transferServer := setupTransferServer(logger, cfg)
	transferrer := setupTransferrer(logger)
	sched := setupScheduler(logger)
	apiRegistry := setupAPIRegistry()
	dsptchr := setupDispatcher(
		logger, sched, transferServer, transferrer, apiRegistry,
	)
	createTransferTasks(logger, cfg, dsptchr)
	apiServer := setupAPIServer(cfg, apiRegistry, dsptchr)

	sigTermCh := make(chan os.Signal)
	signal.Notify(sigTermCh, os.Interrupt)
	signal.Notify(sigTermCh, syscall.SIGTERM)
	go func() {
		<-sigTermCh

		logger.Debug("Closing transfer server...")
		transferServer.Close()

		logger.Debug("Closing scheduler...")
		sched.Stop()

		if apiServer != nil {
			logger.Debug("Closing API server...")
			apiServer.Close()
		}

		logger.Info("Exitting clique-agent...")
	}()

	logger.Info("Clique Agent")

	wg := new(sync.WaitGroup)

	wg.Add(1)
	go func() {
		transferServer.Serve()
		logger.Debug("Transfer server is done.")
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		sched.Run()
		logger.Debug("Scheduler is done.")
		wg.Done()
	}()

	if apiServer != nil {
		wg.Add(1)
		go func() {
			apiServer.Serve()
			logger.Debug("API server is done.")
			wg.Done()
		}()
	}

	wg.Wait()
	logger.Debug("Clique agent is done.")
}

func setupTransferServer(logger *logrus.Logger, cfg config.Config) *transfer.Server {
	server, err := transfer.NewServer(logger, cfg.TransferPort)
	if err != nil {
		logger.Fatalf("Setting up transfer server: %s", err.Error())
	}

	return server
}

func setupTransferrer(logger *logrus.Logger) *transfer.Transferrer {
	return &transfer.Transferrer{Logger: logger}
}

func setupScheduler(logger *logrus.Logger) *scheduler.Scheduler {
	return scheduler.NewScheduler(
		logger,
		// scheduling algorithm
		&scheduler.LotteryTaskSelector{Rand: scheduler.NewCryptoUIG()},
		// sleep between tasks
		time.Second,
		clock.NewClock(),
	)
}

func setupAPIRegistry() *registry.Registry {
	return registry.NewRegistry()
}

func setupDispatcher(
	logger *logrus.Logger,
	scheduler *scheduler.Scheduler,
	transferServer *transfer.Server,
	transferrer *transfer.Transferrer,
	apiRegistry *registry.Registry,
) *dispatcher.Dispatcher {
	return &dispatcher.Dispatcher{
		Scheduler:      scheduler,
		TransferServer: transferServer,
		Transferrer:    transferrer,
		ApiRegistry:    apiRegistry,
		Logger:         logger,
	}
}

func createTransferTasks(
	logger *logrus.Logger,
	cfg config.Config,
	dsptchr *dispatcher.Dispatcher,
) {
	if len(cfg.RemoteHosts) == 0 {
		return
	}

	for _, remoteHost := range cfg.RemoteHosts {
		host, portStr, err := net.SplitHostPort(remoteHost)
		if err != nil {
			logger.Fatalf("Parsing remote host `%s`: %s", remoteHost, err.Error())
		}
		port, err := strconv.ParseInt(portStr, 10, 16)
		if err != nil {
			logger.Fatalf(
				"Parsing remote host's port `%s`: %s",
				remoteHost, err.Error(),
			)
		}

		dsptchr.Create(api.TransferSpec{
			IP:   net.ParseIP(host),
			Port: uint16(port),
			Size: cfg.InitTransferSize,
		})
	}
}

func setupAPIServer(
	cfg config.Config,
	reg *registry.Registry,
	dsptchr *dispatcher.Dispatcher,
) *api.Server {
	if cfg.APIPort == 0 {
		return nil
	}

	return api.NewServer(
		cfg.APIPort,
		reg,
		dsptchr,
	)
}
