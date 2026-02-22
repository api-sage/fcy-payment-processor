package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"
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
	startupStart := time.Now()
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

	// Initialize repositories in parallel
	var wg sync.WaitGroup
	wg.Add(5)

	var userRepoImpl *implementations.UserRepository
	go func() {
		defer wg.Done()
		userRepoImpl = implementations.NewUserRepository(db)
	}()

	var accountRepoImpl *implementations.AccountRepository
	go func() {
		defer wg.Done()
		accountRepoImpl = implementations.NewAccountRepository(db)
	}()

	var rateRepoImpl *implementations.RateRepository
	go func() {
		defer wg.Done()
		rateRepoImpl = implementations.NewRateRepository(db)
	}()

	var transferRepoImpl *implementations.TransferRepository
	go func() {
		defer wg.Done()
		transferRepoImpl = implementations.NewTransferRepository(db)
	}()

	var transientAccountRepoImpl *implementations.TransientAccountRepository
	go func() {
		defer wg.Done()
		transientAccountRepoImpl = implementations.NewTransientAccountRepository(db)
	}()

	wg.Wait()

	participantBankRepo := memory.NewParticipantBankRepository()

	// Ensure default rates before creating services
	if err := rateRepoImpl.EnsureDefaultRates(ctx); err != nil {
		log.Fatalf("ensure default rates: %v", err)
	}

	// Initialize services and controllers in parallel where possible
	var wg2 sync.WaitGroup
	wg2.Add(6)

	var userService *services.UserService
	var userController *controller.UserController
	go func() {
		defer wg2.Done()
		userService = services.NewUserService(userRepoImpl)
		userController = controller.NewUserController(userService)
	}()

	var accountService *services.AccountService
	var accountController *controller.AccountController
	go func() {
		defer wg2.Done()
		accountService = services.NewAccountService(accountRepoImpl, userRepoImpl, participantBankRepo, cfg.GreyBankCode)
		accountController = controller.NewAccountController(accountService)
	}()

	var participantBankService *services.ParticipantBankService
	var participantBankController *controller.ParticipantBankController
	go func() {
		defer wg2.Done()
		participantBankService = services.NewParticipantBankService(participantBankRepo)
		participantBankController = controller.NewParticipantBankController(participantBankService)
	}()

	var rateService *services.RateService
	var rateController *controller.RateController
	go func() {
		defer wg2.Done()
		rateService = services.NewRateService(rateRepoImpl)
		rateController = controller.NewRateController(rateService)
	}()

	var chargesService *services.ChargesService
	var chargesController *controller.ChargesController
	go func() {
		defer wg2.Done()
		chargesService = services.NewChargesService(
			rateRepoImpl,
			cfg.ChargePercent,
			cfg.VATPercent,
			cfg.ChargeMinAmount,
			cfg.ChargeMaxAmount,
		)
		chargesController = controller.NewChargesController(chargesService)
	}()

	var transferController *controller.TransferController
	go func() {
		defer wg2.Done()
		// Ensure transient accounts are set up first
		if err := transientAccountRepoImpl.EnsureInternalAccounts(
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
			transferRepoImpl,
			accountRepoImpl,
			transientAccountRepoImpl,
			transientAccountTransactionRepo,
			participantBankRepo,
			rateRepoImpl,
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
		transferController = controller.NewTransferController(transferService)
	}()

	wg2.Wait()

	mux := router.New(accountController, userController, participantBankController, rateController, chargesController, transferController, middleware.BasicAuth(cfg.ChannelID, cfg.ChannelKey))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	startupDuration := time.Since(startupStart)
	log.Printf("==================================================")
	log.Printf("Application startup completed")
	log.Printf("Start time: %v", startupStart.Format("2006-01-02 15:04:05.000"))
	log.Printf("End time:   %v", time.Now().Format("2006-01-02 15:04:05.000"))
	log.Printf("Total time: %v", startupDuration)
	log.Printf("==================================================")
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("start http server: %v", err)
	}
}
