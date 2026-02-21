package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/repository/repo_interfaces"
	"github.com/api-sage/ccy-payment-processor/src/internal/commons"
	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
	"github.com/api-sage/ccy-payment-processor/src/internal/logger"
	"github.com/api-sage/ccy-payment-processor/src/internal/usecase/service_interfaces"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

type TransferService struct {
	transferRepo                    repo_interfaces.TransferRepository
	accountRepo                     repo_interfaces.AccountRepository
	transientAccountRepo            repo_interfaces.TransientAccountRepository
	transientAccountTransactionRepo repo_interfaces.TransientAccountTransactionRepository
	rateRepo                        repo_interfaces.RateRepository
	rateService                     service_interfaces.RateService
	chargeService                   service_interfaces.ChargesService
	greyBankCode                    string
	internalTransientAccountNumber  string
	internalChargesAccountNumber    string
	internalVATAccountNumber        string
}

func NewTransferService(
	transferRepo repo_interfaces.TransferRepository,
	accountRepo repo_interfaces.AccountRepository,
	transientAccountRepo repo_interfaces.TransientAccountRepository,
	transientAccountTransactionRepo repo_interfaces.TransientAccountTransactionRepository,
	rateRepo repo_interfaces.RateRepository,
	rateService service_interfaces.RateService,
	chargeService service_interfaces.ChargesService,
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
		rateService:                     rateService,
		chargeService:                   chargeService,
		greyBankCode:                    strings.TrimSpace(greyBankCode),
		internalTransientAccountNumber:  strings.TrimSpace(internalTransientAccountNumber),
		internalChargesAccountNumber:    strings.TrimSpace(internalChargesAccountNumber),
		internalVATAccountNumber:        strings.TrimSpace(internalVATAccountNumber),
	}
}

var transferRefCounter uint32

func (s *TransferService) TransferFunds(ctx context.Context, req models.InternalTransferRequest) (commons.Response[models.InternalTransferResponse], error) {
	logger.Info("transfer service internal transfer request", logger.Fields{
		"payload": logger.SanitizePayload(req),
	})

	if err := req.Validate(); err != nil {
		return commons.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error()), err
	}

	beneficiaryBankCode := strings.TrimSpace(req.BeneficiaryBankCode)
	if beneficiaryBankCode != s.greyBankCode {
		err := fmt.Errorf("beneficiaryBankCode is not internal")
		return commons.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error()), err
	}

	debitAccountNumber := strings.TrimSpace(req.DebitAccountNumber)
	creditAccountNumber := strings.TrimSpace(req.CreditAccountNumber)
	if debitAccountNumber == creditAccountNumber {
		err := fmt.Errorf("debitAccountNumber and creditAccountNumber cannot be the same")
		return commons.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error()), err
	}

	debitCurrency := strings.ToUpper(strings.TrimSpace(req.DebitCurrency))
	creditCurrency := strings.ToUpper(strings.TrimSpace(req.CreditCurrency))
	debitAmount, _ := decimal.NewFromString(strings.TrimSpace(req.DebitAmount))

	debitAccount, err := s.accountRepo.GetByAccountNumber(ctx, debitAccountNumber)
	if err != nil {
		if errors.Is(err, commons.ErrRecordNotFound) {
			return commons.ErrorResponse[models.InternalTransferResponse]("Debit account not found"), err
		}
		return commons.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
	}
	creditAccount, err := s.accountRepo.GetByAccountNumber(ctx, creditAccountNumber)
	if err != nil {
		if errors.Is(err, commons.ErrRecordNotFound) {
			return commons.ErrorResponse[models.InternalTransferResponse]("Credit account not found"), err
		}
		return commons.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
	}

	if debitAccount.Status != domain.AccountStatusActive {
		err := fmt.Errorf("debit account is not active")
		return commons.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error()), err
	}
	if creditAccount.Status != domain.AccountStatusActive {
		err := fmt.Errorf("credit account is not active")
		return commons.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error()), err
	}
	if !strings.EqualFold(strings.TrimSpace(debitAccount.Currency), debitCurrency) {
		err := fmt.Errorf("debit currency does not match debit account currency")
		return commons.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error()), err
	}
	if !strings.EqualFold(strings.TrimSpace(creditAccount.Currency), creditCurrency) {
		err := fmt.Errorf("credit currency does not match credit account currency")
		return commons.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error()), err
	}

	convertedAmount, rateUsed, _, err := s.rateService.ConvertRate(ctx, debitAmount.String(), debitCurrency, creditCurrency)
	_, _, charge, vat, totalDebitAmount, err := s.chargeService.GetCharges(ctx, debitAmount.String(), debitCurrency)
	if err != nil {
		return commons.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
	}

	creditAmount, err := decimal.NewFromString(strings.TrimSpace(convertedAmount))
	if err != nil {
		return commons.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
	}

	chargeAmount, err := decimal.NewFromString(strings.TrimSpace(charge))
	if err != nil {
		return commons.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
	}

	vatAmount, err := decimal.NewFromString(strings.TrimSpace(vat))
	if err != nil {
		return commons.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
	}

	sumTotal, err := decimal.NewFromString(strings.TrimSpace(totalDebitAmount))
	if err != nil {
		return commons.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
	}

	debitAvailable, parseErr := decimal.NewFromString(strings.TrimSpace(debitAccount.AvailableBalance))
	if parseErr != nil {
		return commons.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), parseErr
	}

	if debitAvailable.LessThan(sumTotal) {
		err := commons.ErrInsufficientBalance
		return commons.ErrorResponse[models.InternalTransferResponse]("Insufficient balance", err.Error()), err
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
			FCYRate:              rateUsed,
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
			return commons.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
		}
	}
	if err != nil {
		return commons.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
	}

	postingErr := s.transferRepo.ProcessInternalTransfer(
		ctx,
		debitAccountNumber,
		sumTotal.StringFixed(2),
		s.internalTransientAccountNumber,
		req.DebitAmount,
		creditAccountNumber,
		creditAmount.StringFixed(2),
	)
	if postingErr != nil {
		_ = s.transferRepo.UpdateStatus(ctx, createdTransfer.ID, domain.TransferStatusFailed)
		return commons.ErrorResponse[models.InternalTransferResponse]("transfer failed", "Unable to complete transfer posting"), postingErr
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
		return commons.SuccessResponse("Transaction successful. Settlement pending", response), nil
	}

	settlementErr := s.transientAccountRepo.SettleFromSuspenseToFees(
		ctx,
		s.internalTransientAccountNumber,
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
		return commons.SuccessResponse("Transaction successful. Settlement pending", response), nil
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
	return commons.SuccessResponse("Transaction successful", response), nil
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

	rateResp, err := s.rateService.GetRate(ctx, models.GetRateRequest{
		FromCurrency: debitCurrency,
		ToCurrency:   "USD",
	})
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}

	if !rateResp.Success || rateResp.Data == nil {
		return decimal.Zero, decimal.Zero, fmt.Errorf("rate lookup failed")
	}

	rateValue, parseErr := decimal.NewFromString(strings.TrimSpace(rateResp.Data.Rate))
	if parseErr != nil {
		return decimal.Zero, decimal.Zero, fmt.Errorf("invalid USD rate: %w", parseErr)
	}
	if rateValue.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, decimal.Zero, fmt.Errorf("usd rate must be greater than zero")
	}

	chargeUSD := chargeAmount.Mul(rateValue)
	vatUSD := vatAmount.Mul(rateValue)

	return chargeUSD, vatUSD, nil
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
