package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var (
	logger *log.Logger
	config *Config

	configPath = flag.String("config", "", "The configuration file path")
)

type Config struct {
}

func main() {
	logger = log.New(os.Stdout, "clique-agent", 0)
	flag.Parse()

	if *configPath == "" {
		exit("`-config` option is required", 1)
	}

	configContents, err := ioutil.ReadFile(*configPath)
	if err != nil {
		exit(fmt.Sprintf("failed to open file '%s': %s", *configPath, err), 1)
	}

	config = new(Config)
	if err := json.Unmarshal(configContents, &config); err != nil {
		exit(fmt.Sprintf("failed to parse file '%s': %s", *configPath, err), 1)
	}

	log.Print("iCE Clique Agent")
}

func exit(msg string, exitCode int) {
	logger.Fatal(msg)
	os.Exit(exitCode)
}
