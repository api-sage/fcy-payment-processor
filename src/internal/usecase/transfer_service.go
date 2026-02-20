package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
	"github.com/api-sage/ccy-payment-processor/src/internal/logger"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

type TransferService struct {
	transferRepo                    domain.TransferRepository
	accountRepo                     domain.AccountRepository
	transientAccountRepo            domain.TransientAccountRepository
	transientAccountTransactionRepo domain.TransientAccountTransactionRepository
	rateRepo                        domain.RateRepository
	chargePercent                   float64
	vatPercent                      float64
	greyBankCode                    string
	internalTransientAccountNumber  string
	internalChargesAccountNumber    string
	internalVATAccountNumber        string
}

func NewTransferService(
	transferRepo domain.TransferRepository,
	accountRepo domain.AccountRepository,
	transientAccountRepo domain.TransientAccountRepository,
	transientAccountTransactionRepo domain.TransientAccountTransactionRepository,
	rateRepo domain.RateRepository,
	chargePercent float64,
	vatPercent float64,
	greyBankCode string,
	internalTransientAccountNumber string,
	internalChargesAccountNumber string,
	internalVATAccountNumber string,
) *TransferService {
	return &TransferService{
		transferRepo:                    transferRepo,
		accountRepo:                     accountRepo,
		transientAccountRepo:            transientAccountRepo,
		transientAccountTransactionRepo: transientAccountTransactionRepo,
		rateRepo:                        rateRepo,
		chargePercent:                   chargePercent,
		vatPercent:                      vatPercent,
		greyBankCode:                    strings.TrimSpace(greyBankCode),
		internalTransientAccountNumber:  strings.TrimSpace(internalTransientAccountNumber),
		internalChargesAccountNumber:    strings.TrimSpace(internalChargesAccountNumber),
		internalVATAccountNumber:        strings.TrimSpace(internalVATAccountNumber),
	}
}

var transferRefCounter uint32

func (s *TransferService) TransferFunds(ctx context.Context, req models.InternalTransferRequest) (models.Response[models.InternalTransferResponse], error) {
	logger.Info("transfer service internal transfer request", logger.Fields{
		"payload": logger.SanitizePayload(req),
	})

	if err := req.Validate(); err != nil {
		return models.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error()), err
	}

	beneficiaryBankCode := strings.TrimSpace(req.BeneficiaryBankCode)
	if beneficiaryBankCode != s.greyBankCode {
		err := fmt.Errorf("beneficiaryBankCode is not internal")
		return models.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error()), err
	}

	debitAccountNumber := strings.TrimSpace(req.DebitAccountNumber)
	creditAccountNumber := strings.TrimSpace(req.CreditAccountNumber)
	if debitAccountNumber == creditAccountNumber {
		err := fmt.Errorf("debitAccountNumber and creditAccountNumber cannot be the same")
		return models.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error()), err
	}

	debitCurrency := strings.ToUpper(strings.TrimSpace(req.DebitCurrency))
	creditCurrency := strings.ToUpper(strings.TrimSpace(req.CreditCurrency))
	debitAmount, _ := decimal.NewFromString(strings.TrimSpace(req.DebitAmount))

	debitAccount, err := s.accountRepo.GetByAccountNumber(ctx, debitAccountNumber)
	if err != nil {
		if errors.Is(err, domain.ErrRecordNotFound) {
			return models.ErrorResponse[models.InternalTransferResponse]("Debit account not found"), err
		}
		return models.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
	}
	creditAccount, err := s.accountRepo.GetByAccountNumber(ctx, creditAccountNumber)
	if err != nil {
		if errors.Is(err, domain.ErrRecordNotFound) {
			return models.ErrorResponse[models.InternalTransferResponse]("Credit account not found"), err
		}
		return models.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
	}

	if debitAccount.Status != domain.AccountStatusActive {
		err := fmt.Errorf("debit account is not active")
		return models.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error()), err
	}
	if creditAccount.Status != domain.AccountStatusActive {
		err := fmt.Errorf("credit account is not active")
		return models.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error()), err
	}
	if !strings.EqualFold(strings.TrimSpace(debitAccount.Currency), debitCurrency) {
		err := fmt.Errorf("debit currency does not match debit account currency")
		return models.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error()), err
	}
	if !strings.EqualFold(strings.TrimSpace(creditAccount.Currency), creditCurrency) {
		err := fmt.Errorf("credit currency does not match credit account currency")
		return models.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error()), err
	}

	rateValue, _, err := s.resolveRate(ctx, debitCurrency, creditCurrency)
	if err != nil {
		if errors.Is(err, domain.ErrRecordNotFound) {
			return models.ErrorResponse[models.InternalTransferResponse]("Rate not found"), err
		}
		return models.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
	}

	creditAmount := debitAmount.Mul(rateValue)
	chargePercent := decimal.NewFromFloat(s.chargePercent).Div(decimal.NewFromInt(100))
	vatPercent := decimal.NewFromFloat(s.vatPercent).Div(decimal.NewFromInt(100))
	chargeAmount := debitAmount.Mul(chargePercent)
	vatAmount := debitAmount.Mul(vatPercent)
	sumTotal := debitAmount.Add(chargeAmount).Add(vatAmount)

	debitAvailable, parseErr := decimal.NewFromString(strings.TrimSpace(debitAccount.AvailableBalance))
	if parseErr != nil {
		return models.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), parseErr
	}
	if debitAvailable.LessThan(sumTotal) {
		err := domain.ErrInsufficientBalance
		return models.ErrorResponse[models.InternalTransferResponse]("Insufficient balance"), err
	}

	narration := strings.TrimSpace(req.Narration)
	auditPayloadBytes, _ := json.Marshal(logger.SanitizePayload(req))
	auditPayload := string(auditPayloadBytes)

	var createdTransfer domain.Transfer
	for attempt := 0; attempt < 5; attempt++ {
		reference := generateThirtyDigitTransferReference()
		transferRecord := domain.Transfer{
			ExternalRefernece:    stringPtr(reference),
			TransactionReference: stringPtr(reference),
			DebitAccountNumber:   debitAccountNumber,
			CreditAccountNumber:  stringPtr(creditAccountNumber),
			BeneficiaryBankCode:  stringPtr(beneficiaryBankCode),
			DebitCurrency:        debitCurrency,
			CreditCurrency:       creditCurrency,
			DebitAmount:          debitAmount.StringFixed(2),
			CreditAmount:         creditAmount.StringFixed(2),
			FCYRate:              rateValue.StringFixed(8),
			ChargeAmount:         chargeAmount.StringFixed(2),
			VATAmount:            vatAmount.StringFixed(2),
			Narration:            stringPtr(narration),
			Status:               domain.TransferStatusPending,
			AuditPayload:         auditPayload,
		}

		createdTransfer, err = s.transferRepo.Create(ctx, transferRecord)
		if err == nil {
			break
		}
		if !isUniqueViolation(err) {
			return models.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
		}
	}
	if err != nil {
		return models.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
	}

	postingErr := s.transferRepo.ProcessInternalTransfer(
		ctx,
		debitAccountNumber,
		sumTotal.StringFixed(2),
		s.internalTransientAccountNumber,
		creditAccountNumber,
		creditAmount.StringFixed(2),
	)
	if postingErr != nil {
		_ = s.transferRepo.UpdateStatus(ctx, createdTransfer.ID, domain.TransferStatusFailed)
		return models.ErrorResponse[models.InternalTransferResponse]("transfer failed", "Unable to complete transfer posting"), postingErr
	}

	_, _ = s.transientAccountTransactionRepo.Create(ctx, domain.TransientAccountTransaction{
		TransferID:        createdTransfer.ID,
		ExternalRefernece: valueOrEmpty(createdTransfer.TransactionReference),
		EntryType:         domain.LedgerEntryCredit,
		Currency:          debitCurrency,
		Amount:            sumTotal.StringFixed(2),
	})
	_, _ = s.transientAccountTransactionRepo.Create(ctx, domain.TransientAccountTransaction{
		TransferID:        createdTransfer.ID,
		ExternalRefernece: valueOrEmpty(createdTransfer.TransactionReference),
		EntryType:         domain.LedgerEntryDebit,
		Currency:          creditCurrency,
		Amount:            creditAmount.StringFixed(2),
	})

	_ = s.transferRepo.UpdateStatus(ctx, createdTransfer.ID, domain.TransferStatusSuccess)
	createdTransfer.Status = domain.TransferStatusSuccess

	chargeUSD, vatUSD, err := s.convertFeesToUSD(ctx, chargeAmount, vatAmount, debitCurrency)
	if err != nil {
		logger.Error("transfer service convert settlement fees to usd failed", err, logger.Fields{
			"transferId": createdTransfer.ID,
		})
		response := mapTransferToResponse(createdTransfer, sumTotal.StringFixed(2))
		return models.SuccessResponse("Transaction successful. Settlement pending", response), nil
	}

	settlementErr := s.transientAccountRepo.SettleFromSuspenseToFees(
		ctx,
		s.internalTransientAccountNumber,
		debitCurrency,
		chargeAmount.StringFixed(2),
		vatAmount.StringFixed(2),
		s.internalChargesAccountNumber,
		s.internalVATAccountNumber,
		chargeUSD.StringFixed(2),
		vatUSD.StringFixed(2),
	)
	if settlementErr != nil {
		logger.Error("transfer service settlement failed", settlementErr, logger.Fields{
			"transferId": createdTransfer.ID,
		})
		response := mapTransferToResponse(createdTransfer, sumTotal.StringFixed(2))
		return models.SuccessResponse("Transaction successful. Settlement pending", response), nil
	}

	_, _ = s.transientAccountTransactionRepo.Create(ctx, domain.TransientAccountTransaction{
		TransferID:        createdTransfer.ID,
		ExternalRefernece: valueOrEmpty(createdTransfer.TransactionReference),
		EntryType:         domain.LedgerEntryDebit,
		Currency:          debitCurrency,
		Amount:            chargeAmount.StringFixed(2),
	})
	_, _ = s.transientAccountTransactionRepo.Create(ctx, domain.TransientAccountTransaction{
		TransferID:        createdTransfer.ID,
		ExternalRefernece: valueOrEmpty(createdTransfer.TransactionReference),
		EntryType:         domain.LedgerEntryDebit,
		Currency:          debitCurrency,
		Amount:            vatAmount.StringFixed(2),
	})
	_, _ = s.transientAccountTransactionRepo.Create(ctx, domain.TransientAccountTransaction{
		TransferID:        createdTransfer.ID,
		ExternalRefernece: valueOrEmpty(createdTransfer.TransactionReference),
		EntryType:         domain.LedgerEntryCredit,
		Currency:          "USD",
		Amount:            chargeUSD.StringFixed(2),
	})
	_, _ = s.transientAccountTransactionRepo.Create(ctx, domain.TransientAccountTransaction{
		TransferID:        createdTransfer.ID,
		ExternalRefernece: valueOrEmpty(createdTransfer.TransactionReference),
		EntryType:         domain.LedgerEntryCredit,
		Currency:          "USD",
		Amount:            vatUSD.StringFixed(2),
	})

	_ = s.transferRepo.UpdateStatus(ctx, createdTransfer.ID, domain.TransferStatusClosed)
	createdTransfer.Status = domain.TransferStatusClosed

	response := mapTransferToResponse(createdTransfer, sumTotal.StringFixed(2))
	return models.SuccessResponse("Transaction successful", response), nil
}

func (s *TransferService) resolveRate(ctx context.Context, fromCurrency string, toCurrency string) (decimal.Decimal, time.Time, error) {
	if strings.EqualFold(fromCurrency, toCurrency) {
		return decimal.NewFromInt(1), time.Now(), nil
	}

	rate, err := s.rateRepo.GetRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		return decimal.Zero, time.Time{}, err
	}

	value, parseErr := decimal.NewFromString(strings.TrimSpace(rate.Rate))
	if parseErr != nil {
		return decimal.Zero, time.Time{}, parseErr
	}
	if value.Equal(decimal.Zero) {
		return decimal.Zero, time.Time{}, fmt.Errorf("rate cannot be zero")
	}
	return value, rate.RateDate, nil
}

func (s *TransferService) convertFeesToUSD(
	ctx context.Context,
	chargeAmount decimal.Decimal,
	vatAmount decimal.Decimal,
	debitCurrency string,
) (decimal.Decimal, decimal.Decimal, error) {
	if strings.EqualFold(debitCurrency, "USD") {
		return chargeAmount, vatAmount, nil
	}

	rateToUSD, _, err := s.resolveRate(ctx, debitCurrency, "USD")
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	return chargeAmount.Mul(rateToUSD), vatAmount.Mul(rateToUSD), nil
}

func mapTransferToResponse(transfer domain.Transfer, sumTotal string) models.InternalTransferResponse {
	return models.InternalTransferResponse{
		TransactionReference: valueOrEmpty(transfer.TransactionReference),
		DebitAccountNumber:   transfer.DebitAccountNumber,
		CreditAccountNumber:  valueOrEmpty(transfer.CreditAccountNumber),
		BeneficiaryBankCode:  valueOrEmpty(transfer.BeneficiaryBankCode),
		DebitCurrency:        transfer.DebitCurrency,
		CreditCurrency:       transfer.CreditCurrency,
		DebitAmount:          transfer.DebitAmount,
		CreditAmount:         transfer.CreditAmount,
		FcyRate:              transfer.FCYRate,
		ChargeAmount:         transfer.ChargeAmount,
		VATAmount:            transfer.VATAmount,
		SumTotalDebit:        sumTotal,
		Narration:            valueOrEmpty(transfer.Narration),
		Status:               string(transfer.Status),
	}
}

func generateThirtyDigitTransferReference() string {
	now := time.Now().UTC()
	base := now.Format("20060102150405") + fmt.Sprintf("%09d", now.Nanosecond())
	counter := atomic.AddUint32(&transferRefCounter, 1) % 10000000
	suffix := fmt.Sprintf("%07d", counter)
	return base + suffix
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return string(pqErr.Code) == "23505"
	}
	return false
}

func stringPtr(value string) *string {
	v := strings.TrimSpace(value)
	return &v
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
