package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/glestaris/ice-clique/config"
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
	_, err := parseConfig(*configPath)
	if err != nil {
		exit(err.Error(), 1)
	}

	log.Print("iCE Clique Agent")
}

func exit(msg string, exitCode int) {
	logger.Fatal(msg)
	os.Exit(exitCode)
}

func parseConfig(configPath string) (config.Config, error) {
	configContents, err := ioutil.ReadFile(configPath)
	if err != nil {
		return config.Config{}, fmt.Errorf("failed to open file '%s': %s", configPath, err)
	}

	var cfg config.Config
	if err := json.Unmarshal(configContents, &cfg); err != nil {
		return config.Config{}, fmt.Errorf("failed to parse file '%s': %s", configPath, err)
	}

	if err := validateConfig(cfg); err != nil {
		return config.Config{}, fmt.Errorf(
			"invalid configuration file '%s': %s", configPath, err,
		)
	}

	return cfg, nil
}

func validateConfig(cfg config.Config) error {
	return nil
}
