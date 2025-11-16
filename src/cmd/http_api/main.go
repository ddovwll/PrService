package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"PrService/src/cmd/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	app, err := NewApp(cfg)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}

	if err := app.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "app stopped with error: %v\n", err)
		os.Exit(1)
	}
}
