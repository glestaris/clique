// +build !withIperf

package main

import (
	"errors"

	"github.com/Sirupsen/logrus"
	"github.com/ice-stuff/clique/config"
)

func setupIperfTransferrer(
	logger *logrus.Logger, cfg config.Config,
) (transferrer, error) {
	return transferrer{}, errors.New(
		"clique-agent is not compiled with Iperf suppport",
	)
}
