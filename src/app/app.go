package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/controller"
	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/middleware"
	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/router"
	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/repository/implementations"
	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/repository/memory"
	"github.com/api-sage/fcy-payment-processor/src/internal/config"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/services"
	"github.com/api-sage/fcy-payment-processor/src/migrations"
	"github.com/shopspring/decimal"
)

// New loads configuration, runs migrations, and wires the HTTP handler.
// It is shared by the local server (src/cmd/server) and the Vercel entrypoint (api/).
func New(ctx context.Context) (http.Handler, error) {
	decimal.MarshalJSONWithoutQuotes = true

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	migrateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := implementations.RunMigrations(migrateCtx, cfg.DatabaseDSN, migrations.FS); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	db, err := implementations.Open(migrateCtx, cfg.DatabaseDSN, cfg.DBMaxOpenConns, cfg.DBMaxIdleConns)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	userRepo := implementations.NewUserRepository(db)
	accountRepo := implementations.NewAccountRepository(db)
	rateRepo := implementations.NewRateRepository(db)
	transferRepo := implementations.NewTransferRepository(db)
	transientAccountRepo := implementations.NewTransientAccountRepository(db)
	transientAccountTransactionRepo := implementations.NewTransientAccountTransactionRepository(db)
	participantBankRepo := memory.NewParticipantBankRepository()

	if err := rateRepo.EnsureDefaultRates(migrateCtx); err != nil {
		return nil, fmt.Errorf("ensure default rates: %w", err)
	}

	if err := transientAccountRepo.EnsureInternalAccounts(
		migrateCtx,
		cfg.InternalTransientAccountNumber,
		cfg.InternalChargesAccountNumber,
		cfg.InternalVATAccountNumber,
		cfg.ExternalUSDGLAccountNumber,
		cfg.ExternalGBPGLAccountNumber,
		cfg.ExternalEURGLAccountNumber,
		cfg.ExternalNGNGLAccountNumber,
	); err != nil {
		return nil, fmt.Errorf("ensure transient accounts: %w", err)
	}

	userService := services.NewUserService(userRepo)
	userController := controller.NewUserController(userService)

	accountService := services.NewAccountService(accountRepo, userRepo, participantBankRepo, cfg.GreyBankCode)
	accountController := controller.NewAccountController(accountService)

	participantBankService := services.NewParticipantBankService(participantBankRepo)
	participantBankController := controller.NewParticipantBankController(participantBankService)

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

	glTransactionService := services.NewGLTransactionService(transientAccountTransactionRepo)
	glTransactionController := controller.NewGLTransactionController(glTransactionService)

	transactionService := services.NewTransactionService(transientAccountTransactionRepo)
	transactionController := controller.NewTransactionController(transactionService)

	return router.New(
		accountController,
		userController,
		participantBankController,
		rateController,
		chargesController,
		transferController,
		glTransactionController,
		transactionController,
		middleware.BasicAuth(cfg.ChannelID, cfg.ChannelKey),
	), nil
}
