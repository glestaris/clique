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
	"github.com/glestaris/ice-clique/config"
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

	server := setupServer(logger, cfg)

	sched := setupScheduler(logger, cfg)
	for _, t := range transferTasks(logger, cfg, server) {
		sched.Schedule(t)
	}

	sigTermCh := make(chan os.Signal)
	signal.Notify(sigTermCh, os.Interrupt)
	signal.Notify(sigTermCh, syscall.SIGTERM)
	go func() {
		<-sigTermCh
		server.Close()
		sched.Stop()
		logger.Info("Exitting ice-clique...")
	}()

	logger.Info("iCE Clique Agent")

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func() {
		server.Serve()
		wg.Done()
	}()
	go func() {
		sched.Run()
		wg.Done()
	}()

	wg.Wait()
}

func setupServer(logger *logrus.Logger, cfg config.Config) transfer.Server {
	server, err := transfer.NewServer(logger, cfg.TransferPort)
	if err != nil {
		logger.Fatalf("Setting up server: %s", err.Error())
	}

	return server
}

func setupTransferer(
	logger *logrus.Logger,
	cfg config.Config,
) transfer.Transferer {
	return transfer.NewClient(logger)
}

func transferTasks(
	logger *logrus.Logger,
	cfg config.Config,
	server transfer.Server,
) []scheduler.Task {
	if len(cfg.RemoteHosts) == 0 {
		return []scheduler.Task{}
	}

	tasks := make([]scheduler.Task, len(cfg.RemoteHosts))

	for i, remoteHost := range cfg.RemoteHosts {
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

		tasks[i] = &transfer.TransferTask{
			Server:     server,
			Transferer: setupTransferer(logger, cfg),
			TransferSpec: transfer.TransferSpec{
				IP:   net.ParseIP(host),
				Port: uint16(port),
				Size: cfg.InitTransferSize,
			},
			DesiredPriority: 10,
			// logger
			Logger: logger,
		}
	}

	return tasks
}

func setupScheduler(
	logger *logrus.Logger,
	cfg config.Config,
) scheduler.Scheduler {
	return scheduler.NewScheduler(
		// logger
		logger,
		// scheduling algorithm
		&scheduler.LotteryTaskSelector{Rand: scheduler.NewCryptoUIG()},
		// sleep between tasks
		time.Second,
		clock.NewClock(),
	)
}
