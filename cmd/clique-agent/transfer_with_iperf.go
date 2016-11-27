// +build withIperf

package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/ice-stuff/clique/config"
	"github.com/ice-stuff/clique/iperf"
)

func setupIperfTransferrer(
	logger *logrus.Logger, cfg config.Config,
) (transferrer, error) {
	receiver := iperf.NewReceiver(logger, cfg.IperfPort)
	return transferrer{
		interruptible:    receiver,
		transferSender:   iperf.NewSender(logger),
		transferReceiver: receiver,
	}, nil
}
