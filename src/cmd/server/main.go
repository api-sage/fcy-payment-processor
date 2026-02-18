package main

import (
	"context"
	"log"
	"time"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/repository/postgres"
	"github.com/api-sage/ccy-payment-processor/src/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := postgres.RunMigrations(ctx, cfg.DatabaseDSN, cfg.MigrationsDir); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	log.Println("initial migrations completed successfully")
}
