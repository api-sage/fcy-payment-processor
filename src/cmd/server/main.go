package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/controller"
	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/middleware"
	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/router"
	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/repository/memory"
	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/repository/postgres"
	"github.com/api-sage/ccy-payment-processor/src/internal/config"
	"github.com/api-sage/ccy-payment-processor/src/internal/usecase"
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

	db, err := postgres.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	participantBankRepo := memory.NewParticipantBankRepository()

	accountRepo := postgres.NewAccountRepository(db)
	accountService := usecase.NewAccountService(accountRepo, participantBankRepo, cfg.GreyBankCode)
	accountController := controller.NewAccountController(accountService)

	userRepo := postgres.NewUserRepository(db)
	userService := usecase.NewUserService(userRepo)
	userController := controller.NewUserController(userService)

	participantBankService := usecase.NewParticipantBankService(participantBankRepo)
	participantBankController := controller.NewParticipantBankController(participantBankService)

	mux := router.New(accountController, userController, participantBankController, middleware.BasicAuth(cfg.ChannelID, cfg.ChannelKey))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("server listening on %s", addr)
	log.Printf("registered routes: POST /create-account, GET /get-account, POST /create-user, POST /verify-pin, GET /get-participant-banks, GET /swagger")
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("start http server: %v", err)
	}
}
