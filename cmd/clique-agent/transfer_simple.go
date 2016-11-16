package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/ice-stuff/clique/dispatcher"
	"github.com/ice-stuff/clique/transfer"
	"github.com/ice-stuff/clique/transfer/simpletransfer"
)

func setupTransferInterruptible(
	logger *logrus.Logger,
) dispatcher.Interruptible {
	return simpletransfer.NewReceiver(logger)
}

func setupTransferProtocol(logger *logrus.Logger) (
	transfer.TransferReceiver, transfer.TransferSender,
) {
	return simpletransfer.NewReceiver(logger), simpletransfer.NewSender(logger)
}
