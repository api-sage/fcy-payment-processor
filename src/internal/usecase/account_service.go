package usecase

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
)

type AccountService struct {
	accountRepo         domain.AccountRepository
	participantBankRepo domain.ParticipantBankRepository
	greyBankCode        string
}

func NewAccountService(
	accountRepo domain.AccountRepository,
	participantBankRepo domain.ParticipantBankRepository,
	greyBankCode string,
) *AccountService {
	return &AccountService{
		accountRepo:         accountRepo,
		participantBankRepo: participantBankRepo,
		greyBankCode:        strings.TrimSpace(greyBankCode),
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, req models.CreateAccountRequest) (models.Response[models.CreateAccountResponse], error) {
	if err := req.Validate(); err != nil {
		return models.ErrorResponse[models.CreateAccountResponse]("validation failed", err.Error()), err
	}

	balance, err := parseBalance(req.InitialDeposit)
	if err != nil {
		return models.ErrorResponse[models.CreateAccountResponse]("validation failed", err.Error()), err
	}

	account := domain.Account{
		CustomerID:       strings.TrimSpace(req.CustomerID),
		AccountNumber:    generateAccountNumber(),
		Currency:         strings.ToUpper(strings.TrimSpace(req.Currency)),
		AvailableBalance: balance,
		LedgerBalance:    balance,
		Status:           domain.AccountStatusActive,
	}

	created, err := s.accountRepo.Create(ctx, account)
	if err != nil {
		return models.ErrorResponse[models.CreateAccountResponse]("failed to create account", "Unable to create account right now"), err
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

	return models.SuccessResponse("account created successfully", response), nil
}

func (s *AccountService) GetAccount(ctx context.Context, accountNumber string, bankCode string) (models.Response[models.GetAccountResponse], error) {
	accountNumber = strings.TrimSpace(accountNumber)
	bankCode = strings.TrimSpace(bankCode)

	if accountNumber == "" {
		return models.ErrorResponse[models.GetAccountResponse]("validation failed", "accountNumber is required"), fmt.Errorf("accountNumber is required")
	}
	if !isTenDigitAccountNumber(accountNumber) {
		return models.ErrorResponse[models.GetAccountResponse]("validation failed", "accountNumber must be exactly 10 digits"), fmt.Errorf("accountNumber must be exactly 10 digits")
	}
	if bankCode == "" {
		return models.ErrorResponse[models.GetAccountResponse]("validation failed", "bankCode is required"), fmt.Errorf("bankCode is required")
	}
	if !isSixDigitBankCode(bankCode) {
		return models.ErrorResponse[models.GetAccountResponse]("validation failed", "bankCode must be exactly 6 digits"), fmt.Errorf("bankCode must be exactly 6 digits")
	}

	if bankCode != s.greyBankCode {
		banks, err := s.participantBankRepo.GetAll(ctx)
		if err != nil {
			return models.ErrorResponse[models.GetAccountResponse]("failed to get account", "Unable to fetch account right now"), err
		}

		var mappedBankName string
		for _, bank := range banks {
			if strings.TrimSpace(bank.BankCode) == bankCode {
				mappedBankName = strings.TrimSpace(bank.BankName)
				break
			}
		}
		if mappedBankName == "" {
			return models.ErrorResponse[models.GetAccountResponse]("validation failed", "bankCode is not supported"), fmt.Errorf("bankCode is not supported")
		}

		response := models.GetAccountResponse{
			AccountName:   "John III Party",
			AccountNumber: accountNumber,
			BankCode:      bankCode,
			BankName:      mappedBankName,
		}

		return models.SuccessResponse("external account fetched successfully", response), nil
	}

	account, err := s.accountRepo.GetByAccountNumber(ctx, accountNumber)
	if err != nil {
		if errors.Is(err, domain.ErrRecordNotFound) {
			return models.ErrorResponse[models.GetAccountResponse]("Account not found"), err
		}
		return models.ErrorResponse[models.GetAccountResponse]("failed to get account", "Unable to fetch account right now"), err
	}

	response := models.GetAccountResponse{
		ID:               account.ID,
		CustomerID:       account.CustomerID,
		AccountName:      account.CustomerID,
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

	return models.SuccessResponse("account fetched successfully", response), nil
}

func parseBalance(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "0.00", nil
	}

	parsed, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return "", fmt.Errorf("initialDeposit must be a valid number: %w", err)
	}

	if parsed < 0 {
		return "", fmt.Errorf("initialDeposit cannot be negative")
	}

	return fmt.Sprintf("%.2f", parsed), nil
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
