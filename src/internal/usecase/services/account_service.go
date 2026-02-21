package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/repository/repo_interfaces"
	"github.com/api-sage/fcy-payment-processor/src/internal/commons"
	"github.com/api-sage/fcy-payment-processor/src/internal/domain"
	"github.com/api-sage/fcy-payment-processor/src/internal/logger"
	"github.com/shopspring/decimal"
)

type AccountService struct {
	accountRepo         repo_interfaces.AccountRepository
	userRepo            domain.UserRepository
	participantBankRepo domain.ParticipantBankRepository
	greyBankCode        string
}

func NewAccountService(
	accountRepo repo_interfaces.AccountRepository,
	userRepo domain.UserRepository,
	participantBankRepo domain.ParticipantBankRepository,
	greyBankCode string,
) *AccountService {
	return &AccountService{
		accountRepo:         accountRepo,
		userRepo:            userRepo,
		participantBankRepo: participantBankRepo,
		greyBankCode:        strings.TrimSpace(greyBankCode),
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, req models.CreateAccountRequest) (commons.Response[models.CreateAccountResponse], error) {
	logger.Info("account service create account request", logger.Fields{
		"payload": logger.SanitizePayload(req),
	})

	if err := req.Validate(); err != nil {
		logger.Error("account service create account validation failed", err, nil)
		return commons.ErrorResponse[models.CreateAccountResponse]("validation failed", err.Error()), err
	}

	customerID := strings.TrimSpace(req.CustomerID)
	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	hasAccount, err := s.accountRepo.HasAccountForCustomerIDAndCurrency(ctx, customerID, currency)
	if err != nil {
		logger.Error("account service create account existing account check failed", err, logger.Fields{
			"customerId": customerID,
			"currency":   currency,
		})
		return commons.ErrorResponse[models.CreateAccountResponse]("failed to create account", "Customer already has an account for this currency"), err
	}
	if hasAccount {
		err := fmt.Errorf("customerId already has a %s account", currency)
		return commons.ErrorResponse[models.CreateAccountResponse]("validation failed", err.Error()), err
	}

	balance, err := parseBalance(req.InitialDeposit)
	if err != nil {
		logger.Error("account service create account parse balance failed", err, nil)
		return commons.ErrorResponse[models.CreateAccountResponse]("validation failed", err.Error()), err
	}

	account := domain.Account{
		CustomerID:       customerID,
		AccountNumber:    generateAccountNumber(),
		Currency:         currency,
		AvailableBalance: balance,
		LedgerBalance:    balance,
		Status:           domain.AccountStatusActive,
	}

	created, err := s.accountRepo.Create(ctx, account)
	if err != nil {
		logger.Error("account service create account repository failed", err, logger.Fields{
			"customerId": account.CustomerID,
		})
		return commons.ErrorResponse[models.CreateAccountResponse]("failed to create account", "Unable to create account right now"), err
	}

	response := models.CreateAccountResponse{
		ID:               created.ID,
		CustomerID:       created.CustomerID,
		AccountNumber:    created.AccountNumber,
		Currency:         created.Currency,
		AvailableBalance: created.AvailableBalance,
		LedgerBalance:    created.LedgerBalance,
		Status:           string(created.Status),
		CreatedAt:        created.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        created.UpdatedAt.Format(time.RFC3339),
	}

	logger.Info("account service create account success", logger.Fields{
		"accountId":     response.ID,
		"accountNumber": response.AccountNumber,
		"customerId":    response.CustomerID,
	})

	return commons.SuccessResponse("account created successfully", response), nil
}

func (s *AccountService) GetAccount(ctx context.Context, accountNumber string, bankCode string) (commons.Response[models.GetAccountResponse], error) {
	logger.Info("account service get account request", logger.Fields{
		"accountNumber": accountNumber,
		"bankCode":      bankCode,
	})

	accountNumber = strings.TrimSpace(accountNumber)
	bankCode = strings.TrimSpace(bankCode)

	if accountNumber == "" {
		return commons.ErrorResponse[models.GetAccountResponse]("validation failed", "accountNumber is required"), fmt.Errorf("accountNumber is required")
	}
	if !isTenDigitAccountNumber(accountNumber) {
		return commons.ErrorResponse[models.GetAccountResponse]("validation failed", "accountNumber must be exactly 10 digits"), fmt.Errorf("accountNumber must be exactly 10 digits")
	}
	if bankCode == "" {
		return commons.ErrorResponse[models.GetAccountResponse]("validation failed", "bankCode is required"), fmt.Errorf("bankCode is required")
	}
	if !isSixDigitBankCode(bankCode) {
		return commons.ErrorResponse[models.GetAccountResponse]("validation failed", "bankCode must be exactly 6 digits"), fmt.Errorf("bankCode must be exactly 6 digits")
	}

	if bankCode != s.greyBankCode {
		banks, err := s.participantBankRepo.GetAll(ctx)
		if err != nil {
			logger.Error("account service get external account banks lookup failed", err, logger.Fields{
				"bankCode": bankCode,
			})
			return commons.ErrorResponse[models.GetAccountResponse]("failed to get account", "Unable to fetch account right now"), err
		}

		var mappedBankName string
		for _, bank := range banks {
			if strings.TrimSpace(bank.BankCode) == bankCode {
				mappedBankName = strings.TrimSpace(bank.BankName)
				break
			}
		}
		if mappedBankName == "" {
			return commons.ErrorResponse[models.GetAccountResponse]("validation failed", "bankCode is not supported"), fmt.Errorf("bankCode is not supported")
		}

		response := models.GetAccountResponse{
			AccountName:   "John III Party",
			AccountNumber: accountNumber,
			BankCode:      bankCode,
			BankName:      mappedBankName,
		}

		logger.Info("account service get external account success", logger.Fields{
			"accountNumber": accountNumber,
			"bankCode":      bankCode,
			"bankName":      mappedBankName,
		})

		return commons.SuccessResponse("external account fetched successfully", response), nil
	}

	account, err := s.accountRepo.GetByAccountNumber(ctx, accountNumber)
	if err != nil {
		logger.Error("account service get internal account failed", err, logger.Fields{
			"accountNumber": accountNumber,
			"bankCode":      bankCode,
		})
		if errors.Is(err, commons.ErrRecordNotFound) {
			return commons.ErrorResponse[models.GetAccountResponse]("Account not found"), err
		}
		return commons.ErrorResponse[models.GetAccountResponse]("failed to get account", "Unable to fetch account right now"), err
	}

	user, err := s.userRepo.GetByCustomerID(ctx, account.CustomerID)
	if err != nil {
		logger.Error("account service get user for account name failed", err, logger.Fields{
			"accountNumber": accountNumber,
			"customerId":    account.CustomerID,
		})
		if errors.Is(err, commons.ErrRecordNotFound) {
			return commons.ErrorResponse[models.GetAccountResponse]("User not found"), err
		}
		return commons.ErrorResponse[models.GetAccountResponse]("failed to get account", "Unable to fetch account right now"), err
	}

	accountName := strings.TrimSpace(strings.Join([]string{
		strings.TrimSpace(user.FirstName),
		strings.TrimSpace(user.LastName),
	}, " "))

	response := models.GetAccountResponse{
		ID:               account.ID,
		CustomerID:       account.CustomerID,
		AccountName:      accountName,
		AccountNumber:    account.AccountNumber,
		BankCode:         bankCode,
		BankName:         "Grey",
		Currency:         account.Currency,
		AvailableBalance: account.AvailableBalance,
		LedgerBalance:    account.LedgerBalance,
		Status:           string(account.Status),
		CreatedAt:        account.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        account.UpdatedAt.Format(time.RFC3339),
	}

	logger.Info("account service get internal account success", logger.Fields{
		"accountId":     response.ID,
		"accountNumber": response.AccountNumber,
		"customerId":    response.CustomerID,
	})

	return commons.SuccessResponse("account fetched successfully", response), nil
}

func (s *AccountService) DepositFunds(ctx context.Context, req models.DepositFundsRequest) (commons.Response[models.DepositFundsResponse], error) {
	logger.Info("account service deposit funds request", logger.Fields{
		"payload": logger.SanitizePayload(req),
	})

	if err := req.Validate(); err != nil {
		logger.Error("account service deposit funds validation failed", err, nil)
		return commons.ErrorResponse[models.DepositFundsResponse]("validation failed", err.Error()), err
	}

	accountNumber := strings.TrimSpace(req.AccountNumber)
	amount := req.Amount

	if err := s.accountRepo.DepositFunds(ctx, accountNumber, amount); err != nil {
		logger.Error("account service deposit funds failed", err, logger.Fields{
			"accountNumber": accountNumber,
			"amount":        amount,
		})
		if errors.Is(err, commons.ErrRecordNotFound) {
			return commons.ErrorResponse[models.DepositFundsResponse]("Account not found"), err
		}
		return commons.ErrorResponse[models.DepositFundsResponse]("failed to deposit funds", "Unable to deposit funds right now"), err
	}

	account, err := s.accountRepo.GetByAccountNumber(ctx, accountNumber)
	if err != nil {
		logger.Error("account service get account after deposit failed", err, logger.Fields{
			"accountNumber": accountNumber,
		})
		if errors.Is(err, commons.ErrRecordNotFound) {
			return commons.ErrorResponse[models.DepositFundsResponse]("Account not found"), err
		}
		return commons.ErrorResponse[models.DepositFundsResponse]("failed to fetch account", "Unable to fetch account right now"), err
	}

	response := models.DepositFundsResponse{
		AccountNumber:    account.AccountNumber,
		Currency:         account.Currency,
		DepositedAmount:  amount,
		AvailableBalance: account.AvailableBalance,
		LedgerBalance:    account.LedgerBalance,
	}

	logger.Info("account service deposit funds success", logger.Fields{
		"accountNumber":    response.AccountNumber,
		"depositedAmount":  response.DepositedAmount,
		"availableBalance": response.AvailableBalance,
	})

	return commons.SuccessResponse("funds deposited successfully", response), nil
}

func parseBalance(raw *decimal.Decimal) (decimal.Decimal, error) {
	if raw == nil {
		return decimal.Zero, nil
	}
	if raw.LessThan(decimal.Zero) {
		return decimal.Zero, fmt.Errorf("initialDeposit cannot be negative")
	}

	return raw.Round(2), nil
}

func generateAccountNumber() string {
	return fmt.Sprintf("%010d", time.Now().UnixNano()%10_000_000_000)
}

func isTenDigitAccountNumber(accountNumber string) bool {
	if len(accountNumber) != 10 {
		return false
	}

	for _, ch := range accountNumber {
		if ch < '0' || ch > '9' {
			return false
		}
	}

	return true
}

func isSixDigitBankCode(bankCode string) bool {
	if len(bankCode) != 6 {
		return false
	}

	for _, ch := range bankCode {
		if ch < '0' || ch > '9' {
			return false
		}
	}

	return true
}
