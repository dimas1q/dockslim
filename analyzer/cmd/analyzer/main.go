package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dimas1q/dockslim/analyzer/internal/config"
	"github.com/dimas1q/dockslim/analyzer/internal/db"
	"github.com/dimas1q/dockslim/analyzer/internal/worker"
)

func main() {
	cfg := config.Load()

	log.Printf("Analyzer worker started")
	log.Printf("Using Postgres DSN: %s", cfg.PostgresDSN)
	log.Printf("Using Redis address: %s", cfg.RedisAddr)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbConn, err := db.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer dbConn.Close()

	analysisWorker := worker.New(dbConn)
	if err := analysisWorker.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("worker stopped: %v", err)
	}
}
