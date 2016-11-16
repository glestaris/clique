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
	"github.com/ice-stuff/clique/transfer/simpletransfer"
)

var (
	cfg config.Config

	configPath = flag.String("config", "", "The configuration file path")

	version = flag.Bool("version", false, "Print clique-agent version")
	debug   = flag.Bool("debug", false, "Print debug messaging")
)

func main() {
	///// ARGUMENT PARSING //////////////////////////////////////////////////////

	flag.Parse()
	if *version {
		fmt.Printf("clique-agent v%s\n", clique.CliqueAgentVersion)
		os.Exit(0)
	}
	level := logrus.InfoLevel
	if *debug {
		level = logrus.DebugLevel
	}
	if *configPath == "" {
		fmt.Fprintf(os.Stderr, "`-config` option is required\n")
		os.Exit(1)
	}

	///// LOGGING ///////////////////////////////////////////////////////////////

	logger := &logrus.Logger{
		Out:       os.Stdout,
		Level:     level,
		Formatter: new(logrus.TextFormatter),
	}
	logger.Debug("Initializing internals...")

	///// CONFIGURATION /////////////////////////////////////////////////////////

	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		logger.Fatal(err.Error())
	}

	///// TRANSFER //////////////////////////////////////////////////////////////

	// Protocol
	transferReceiver := simpletransfer.NewReceiver(logger)
	transferSender := simpletransfer.NewSender(logger)

	// Server
	transferListener, err := net.Listen(
		"tcp", fmt.Sprintf("0.0.0.0:%d", cfg.TransferPort),
	)
	if err != nil {
		logger.Fatalf("Setting up transfer server: %s", err.Error())
	}
	transferServer := transfer.NewServer(
		logger, transferListener, transferReceiver,
	)

	// Client
	transferConnector := transfer.NewConnector()
	transferClient := transfer.NewClient(
		logger, transferConnector, transferSender,
	)

	///// SCHEDULING ////////////////////////////////////////////////////////////

	schedRandGen := scheduler.NewCryptoUIG()
	schedTaskSelector := &scheduler.LotteryTaskSelector{
		Rand: schedRandGen,
	}
	schedClock := clock.NewClock()
	sched := scheduler.NewScheduler(
		logger,
		schedTaskSelector, // scheduling algorithm
		time.Second,       // sleep between tasks
		schedClock,
	)

	///// TRANSFER REGISTRY /////////////////////////////////////////////////////

	transferRegistry := registry.NewRegistry()

	///// DISPATCHER ////////////////////////////////////////////////////////////

	dsptchr := &dispatcher.Dispatcher{
		Scheduler:      sched,
		TransferServer: transferReceiver,
		Transferrer:    transferClient,
		ApiRegistry:    transferRegistry,
		Logger:         logger,
	}

	///// API ///////////////////////////////////////////////////////////////////

	var apiServer *api.Server
	if cfg.APIPort != 0 {
		apiServer = api.NewServer(
			cfg.APIPort,
			transferRegistry,
			dsptchr,
		)
	}

	///// SIGNAL HANDLER ////////////////////////////////////////////////////////

	sigTermCh := make(chan os.Signal)
	signal.Notify(sigTermCh, os.Interrupt)
	signal.Notify(sigTermCh, syscall.SIGTERM)
	go func() {
		<-sigTermCh

		logger.Debug("Closing transfer server...")
		transferListener.Close()

		logger.Debug("Closing scheduler...")
		sched.Stop()

		if apiServer != nil {
			logger.Debug("Closing API server...")
			apiServer.Close()
		}

		logger.Info("Exitting clique-agent...")
	}()
	logger.Debug("Initialization is complete!")

	///// START /////////////////////////////////////////////////////////////////

	logger.Info("Clique Agent")

	// Populate the dispatcher with tasks
	if len(cfg.RemoteHosts) != 0 {
		createTransferTasks(logger, cfg, dsptchr)
	}

	wg := new(sync.WaitGroup)

	// Start the transfer server
	wg.Add(1)
	go func() {
		transferServer.Serve()
		logger.Debug("Transfer server is done.")
		wg.Done()
	}()

	// Start the scheduler
	wg.Add(1)
	go func() {
		sched.Run()
		logger.Debug("Scheduler is done.")
		wg.Done()
	}()

	// Start the API server
	if apiServer != nil {
		wg.Add(1)
		go func() {
			apiServer.Serve()
			logger.Debug("API server is done.")
			wg.Done()
		}()
	}

	// Wait until everything is done!
	wg.Wait()
	logger.Debug("Clique agent is done.")
}

func createTransferTasks(
	logger *logrus.Logger,
	cfg config.Config,
	dsptchr *dispatcher.Dispatcher,
) {
	logger.Debugf(
		"Found %d remote hosts for the initial transfers", len(cfg.RemoteHosts),
	)
	logger.Debugf("Size of initial transfers = %d bytes", cfg.InitTransferSize)

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
