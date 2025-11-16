// Package main provides HTTP API for PR reviewer assignment.
//
// @title		PR Reviewer Assignment Service (Test Task, Fall 2025)
// @version	1.0.0
package main

import (
	"context"
	"log"
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
		log.Fatalf("app stopped with error: %v\n", err)
	}
}
