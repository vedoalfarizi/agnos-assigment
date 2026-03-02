package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Initialize application bootstrap
	bootstrap := &Bootstrap{}
	if err := bootstrap.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bootstrap application: %v\n", err)
		os.Exit(1)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine to allow signal handling
	serverErrChan := make(chan error, 1)
	go func() {
		addr := bootstrap.GetServerAddr()
		bootstrap.Logger.Infof("Starting server on %s", addr)
		serverErrChan <- bootstrap.Router.Run(addr)
	}()

	// Wait for either server error or shutdown signal
	select {
	case sig := <-sigChan:
		bootstrap.Logger.Infof("Received signal: %v", sig)
		// Gracefully shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if err := bootstrap.Shutdown(ctx); err != nil {
			bootstrap.Logger.Errorf("Shutdown error: %v", err)
			os.Exit(1)
		}
	case err := <-serverErrChan:
		if err != nil {
			bootstrap.Logger.Errorf("Server error: %v", err)
			os.Exit(1)
		}
	}
}
