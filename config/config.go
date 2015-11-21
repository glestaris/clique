package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

type Config struct {
	TransferPort uint16 `json:"transfer_port"`
}

func NewConfig(configPath string) (Config, error) {
	configContents, err := ioutil.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open file '%s': %s", configPath, err)
	}

	var cfg Config
	if err := json.Unmarshal(configContents, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse file '%s': %s", configPath, err)
	}

	if err := validateConfig(cfg); err != nil {
		return Config{}, fmt.Errorf(
			"invalid configuration file '%s': %s", configPath, err,
		)
	}

	return cfg, nil
}

func validateConfig(cfg Config) error {
	if cfg.TransferPort == 0 {
		return errors.New("transfer port is not defined")
	}

	return nil
}
