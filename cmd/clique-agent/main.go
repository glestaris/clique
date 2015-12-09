package main

import (
	"flag"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/glestaris/ice-clique/api"
	"github.com/glestaris/ice-clique/api/registry"
	"github.com/glestaris/ice-clique/config"
	"github.com/glestaris/ice-clique/dispatcher"
	"github.com/glestaris/ice-clique/scheduler"
	"github.com/glestaris/ice-clique/transfer"
	"github.com/pivotal-golang/clock"
)

var (
	cfg config.Config

	configPath = flag.String("config", "", "The configuration file path")
	debug      = flag.Bool("debug", false, "Print debug messaging")
)

func main() {
	flag.Parse()

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
	transferer := setupTransferer(logger)
	sched := setupScheduler(logger)
	apiRegistry := setupAPIRegistry()
	dsptchr := setupDispatcher(
		logger, sched, transferServer, transferer, apiRegistry,
	)
	createTransferTasks(logger, cfg, dsptchr)
	apiServer := setupAPIServer(cfg, apiRegistry, dsptchr)

	sigTermCh := make(chan os.Signal)
	signal.Notify(sigTermCh, os.Interrupt)
	signal.Notify(sigTermCh, syscall.SIGTERM)
	go func() {
		<-sigTermCh
		transferServer.Close()
		sched.Stop()
		if apiServer != nil {
			apiServer.Close()
		}
		logger.Info("Exitting clique-agent...")
	}()

	logger.Info("iCE Clique Agent")

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func() {
		transferServer.Serve()
		wg.Done()
	}()
	go func() {
		sched.Run()
		wg.Done()
	}()
	if apiServer != nil {
		wg.Add(1)
		go func() {
			apiServer.Serve()
			wg.Done()
		}()
	}

	wg.Wait()
}

func setupTransferServer(logger *logrus.Logger, cfg config.Config) *transfer.Server {
	server, err := transfer.NewServer(logger, cfg.TransferPort)
	if err != nil {
		logger.Fatalf("Setting up transfer server: %s", err.Error())
	}

	return server
}

func setupTransferer(logger *logrus.Logger) *transfer.Transferer {
	return &transfer.Transferer{Logger: logger}
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
	transferer *transfer.Transferer,
	apiRegistry *registry.Registry,
) *dispatcher.Dispatcher {
	return &dispatcher.Dispatcher{
		Scheduler:      scheduler,
		TransferServer: transferServer,
		Transferer:     transferer,
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
