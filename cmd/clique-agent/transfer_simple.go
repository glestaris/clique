// +build !transferIsIperf

package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/ice-stuff/clique/dispatcher"
	"github.com/ice-stuff/clique/transfer"
	"github.com/ice-stuff/clique/transfer/simple"
)

func setupTransferInterruptible(
	logger *logrus.Logger,
) dispatcher.Interruptible {
	return simple.NewReceiver(logger)
}

func setupTransferProtocol(logger *logrus.Logger) (
	transfer.TransferReceiver, transfer.TransferSender,
) {
	return simple.NewReceiver(logger), simple.NewSender(logger)
}
