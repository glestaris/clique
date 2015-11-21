package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/glestaris/ice-clique/config"
	"github.com/glestaris/ice-clique/transfer"
)

var (
	logger *log.Logger
	cfg    config.Config

	configPath = flag.String("config", "", "The configuration file path")
)

func main() {
	logger = log.New(os.Stdout, "clique-agent", 0)
	flag.Parse()

	if *configPath == "" {
		exit("`-config` option is required", 1)
	}
	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		exit(err.Error(), 1)
	}

	log.Print("iCE Clique Agent")
	server, err := transfer.NewServer(cfg.TransferPort)
	if err != nil {
		exit(fmt.Sprintf("starting transferring server: %s", err.Error()), 2)
	}

	sigTermCh := make(chan os.Signal)
	signal.Notify(sigTermCh, os.Interrupt)
	signal.Notify(sigTermCh, syscall.SIGTERM)
	go func(c chan os.Signal, server transfer.Server) {
		<-c
		server.Close()
		log.Print("Exitting...")
	}(sigTermCh, server)

	server.Serve()
	os.Exit(0)
}

func exit(msg string, exitCode int) {
	logger.Fatal(msg)
	os.Exit(exitCode)
}
