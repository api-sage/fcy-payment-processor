package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/controller"
	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/middleware"
	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/router"
	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/repository/implementations"
	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/repository/memory"
	"github.com/api-sage/fcy-payment-processor/src/internal/config"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/services"
	"github.com/shopspring/decimal"
)

func main() {
	decimal.MarshalJSONWithoutQuotes = true

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := implementations.RunMigrations(ctx, cfg.DatabaseDSN, cfg.MigrationsDir); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	db, err := implementations.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	participantBankRepo := memory.NewParticipantBankRepository()

	accountRepo := implementations.NewAccountRepository(db)
	accountService := services.NewAccountService(accountRepo, participantBankRepo, cfg.GreyBankCode)
	accountController := controller.NewAccountController(accountService)

	userRepo := implementations.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	userController := controller.NewUserController(userService)

	participantBankService := services.NewParticipantBankService(participantBankRepo)
	participantBankController := controller.NewParticipantBankController(participantBankService)

	rateRepo := implementations.NewRateRepository(db)
	if err := rateRepo.EnsureDefaultRates(ctx); err != nil {
		log.Fatalf("ensure default rates: %v", err)
	}
	rateService := services.NewRateService(rateRepo)
	rateController := controller.NewRateController(rateService)

	chargesService := services.NewChargesService(
		rateRepo,
		cfg.ChargePercent,
		cfg.VATPercent,
		cfg.ChargeMinAmount,
		cfg.ChargeMaxAmount,
	)
	chargesController := controller.NewChargesController(chargesService)

	transferRepo := implementations.NewTransferRepository(db)
	transientAccountRepo := implementations.NewTransientAccountRepository(db)
	if err := transientAccountRepo.EnsureInternalAccounts(
		ctx,
		cfg.InternalTransientAccountNumber,
		cfg.InternalChargesAccountNumber,
		cfg.InternalVATAccountNumber,
		cfg.ExternalUSDGLAccountNumber,
		cfg.ExternalGBPGLAccountNumber,
		cfg.ExternalEURGLAccountNumber,
		cfg.ExternalNGNGLAccountNumber,
	); err != nil {
		log.Fatalf("ensure transient accounts: %v", err)
	}
	transientAccountTransactionRepo := implementations.NewTransientAccountTransactionRepository(db)
	transferService := services.NewTransferService(
		transferRepo,
		accountRepo,
		transientAccountRepo,
		transientAccountTransactionRepo,
		participantBankRepo,
		rateRepo,
		userService,
		rateService,
		chargesService,
		cfg.GreyBankCode,
		cfg.InternalTransientAccountNumber,
		cfg.InternalChargesAccountNumber,
		cfg.InternalVATAccountNumber,
		cfg.ExternalUSDGLAccountNumber,
		cfg.ExternalGBPGLAccountNumber,
		cfg.ExternalEURGLAccountNumber,
		cfg.ExternalNGNGLAccountNumber,
	)
	transferController := controller.NewTransferController(transferService)

	mux := router.New(accountController, userController, participantBankController, rateController, chargesController, transferController, middleware.BasicAuth(cfg.ChannelID, cfg.ChannelKey))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("start http server: %v", err)
	}
}
