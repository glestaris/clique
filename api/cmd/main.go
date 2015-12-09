package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/glestaris/ice-clique/api"
	"github.com/glestaris/ice-clique/api/fakes"
)

func main() {
	terminate := make(chan struct{})

	// initialize the server
	server := api.NewServer(
		6000, // port
		new(fakes.FakeTransferResultsRegistry), // transfer results registry
		new(fakes.FakeTransferCreator),         // transfer creator
	)

	// signal handler
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGTERM)
	signal.Notify(sigChan, syscall.SIGINT)
	go func() {
		for {
			<-sigChan

			fmt.Println("Receieved signal, closing the server")
			if err := server.Close(); err != nil {
				fmt.Println("Failed to close the server:", err)
				close(terminate)
			}
		}
	}()

	// run the server
	fmt.Println("Running the server")
	server.Serve()

	// wait for signal handler
	<-terminate
}
