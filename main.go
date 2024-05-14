package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/jskiba/papertrail-cli-poc/swo"
)

func main() {
	opts, err := swo.NewOptions(os.Args[1:])
	if err != nil {
		slog.Error("Could not prepare options", "error", err)
		os.Exit(1)
	}

	client, err := swo.NewClient(opts)
	if err != nil {
		slog.Error("Could not create a client for communicating with SWO", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	err = client.Run(ctx)
	if err != nil {
		slog.Error("An error occured while trying to communicate with SWO", "error", err)
		os.Exit(1)
	}
}
