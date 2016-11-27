package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/ice-stuff/clique/config"
	"github.com/ice-stuff/clique/dispatcher"
	"github.com/ice-stuff/clique/transfer"
	"github.com/ice-stuff/clique/transfer/simple"
)

type transferrer struct {
	interruptible    dispatcher.Interruptible
	transferReceiver transfer.TransferReceiver
	transferSender   transfer.TransferSender
}

func setupTransferrer(logger *logrus.Logger, cfg config.Config) (
	transferrer, error,
) {
	if cfg.UseIperf {
		return setupIperfTransferrer(logger, cfg)
	}

	return setupSimpleTransferrer(logger, cfg)
}

func setupSimpleTransferrer(
	logger *logrus.Logger, cfg config.Config,
) (transferrer, error) {
	receiver := simple.NewReceiver(logger)
	return transferrer{
		interruptible:    receiver,
		transferSender:   simple.NewSender(logger),
		transferReceiver: receiver,
	}, nil
}
