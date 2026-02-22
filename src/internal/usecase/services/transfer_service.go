package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/repository/repo_interfaces"
	"github.com/api-sage/fcy-payment-processor/src/internal/commons"
	"github.com/api-sage/fcy-payment-processor/src/internal/domain"
	"github.com/api-sage/fcy-payment-processor/src/internal/logger"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/service_interfaces"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

// Verify that TransferService implements the service_interfaces.TransferService interface
var _ service_interfaces.TransferService = (*TransferService)(nil)

type TransferService struct {
	transferRepo                    repo_interfaces.TransferRepository
	accountRepo                     repo_interfaces.AccountRepository
	transientAccountRepo            repo_interfaces.TransientAccountRepository
	transientAccountTransactionRepo repo_interfaces.TransientAccountTransactionRepository
	participantBankRepo             domain.ParticipantBankRepository
	rateRepo                        repo_interfaces.RateRepository
	userService                     service_interfaces.UserService
	rateService                     service_interfaces.RateService
	chargeService                   service_interfaces.ChargesService
	greyBankCode                    string
	internalTransientAccountNumber  string
	internalChargesAccountNumber    string
	internalVATAccountNumber        string
	externalUSDGLAccountNumber      string
	externalGBPGLAccountNumber      string
	externalEURGLAccountNumber      string
	externalNGNGLAccountNumber      string
}

func NewTransferService(
	transferRepo repo_interfaces.TransferRepository,
	accountRepo repo_interfaces.AccountRepository,
	transientAccountRepo repo_interfaces.TransientAccountRepository,
	transientAccountTransactionRepo repo_interfaces.TransientAccountTransactionRepository,
	participantBankRepo domain.ParticipantBankRepository,
	rateRepo repo_interfaces.RateRepository,
	userService service_interfaces.UserService,
	rateService service_interfaces.RateService,
	chargeService service_interfaces.ChargesService,
	greyBankCode string,
	internalTransientAccountNumber string,
	internalChargesAccountNumber string,
	internalVATAccountNumber string,
	externalUSDGLAccountNumber string,
	externalGBPGLAccountNumber string,
	externalEURGLAccountNumber string,
	externalNGNGLAccountNumber string,
) *TransferService {
	return &TransferService{
		transferRepo:                    transferRepo,
		accountRepo:                     accountRepo,
		transientAccountRepo:            transientAccountRepo,
		transientAccountTransactionRepo: transientAccountTransactionRepo,
		participantBankRepo:             participantBankRepo,
		rateRepo:                        rateRepo,
		userService:                     userService,
		rateService:                     rateService,
		chargeService:                   chargeService,
		greyBankCode:                    strings.TrimSpace(greyBankCode),
		internalTransientAccountNumber:  strings.TrimSpace(internalTransientAccountNumber),
		internalChargesAccountNumber:    strings.TrimSpace(internalChargesAccountNumber),
		internalVATAccountNumber:        strings.TrimSpace(internalVATAccountNumber),
		externalUSDGLAccountNumber:      strings.TrimSpace(externalUSDGLAccountNumber),
		externalGBPGLAccountNumber:      strings.TrimSpace(externalGBPGLAccountNumber),
		externalEURGLAccountNumber:      strings.TrimSpace(externalEURGLAccountNumber),
		externalNGNGLAccountNumber:      strings.TrimSpace(externalNGNGLAccountNumber),
	}
}

var transferRefCounter uint32

func (s *TransferService) TransferFunds(ctx context.Context, req models.InternalTransferRequest) (commons.Response[models.InternalTransferResponse], error) {
	logger.Info("transfer service transfer request", logger.Fields{
		"payload": logger.SanitizePayload(req),
	})

	if err := req.Validate(); err != nil {
		return commons.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error()), err
	}

	beneficiaryBankCode := strings.TrimSpace(req.BeneficiaryBankCode)
	if beneficiaryBankCode != s.greyBankCode {
		return s.processExternalTransfer(ctx, req)
	}

	debitAccountNumber := strings.TrimSpace(req.DebitAccountNumber)
	creditAccountNumber := strings.TrimSpace(req.CreditAccountNumber)
	if debitAccountNumber == creditAccountNumber {
		err := fmt.Errorf("debitAccountNumber and creditAccountNumber cannot be the same")
		return commons.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error()), err
	}

	debitCurrency := strings.ToUpper(strings.TrimSpace(req.DebitCurrency))
	creditCurrency := strings.ToUpper(strings.TrimSpace(req.CreditCurrency))
	debitAmount := req.DebitAmount

	// Fetch both accounts in parallel
	var debitAccount, creditAccount *domain.Account
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		var err error
		debitAccount, err = s.accountRepo.GetByAccountNumber(egCtx, debitAccountNumber)
		if err != nil {
			if errors.Is(err, commons.ErrRecordNotFound) {
				return fmt.Errorf("debit account not found")
			}
			return fmt.Errorf("failed to fetch debit account")
		}
		return nil
	})

	eg.Go(func() error {
		var err error
		creditAccount, err = s.accountRepo.GetByAccountNumber(egCtx, creditAccountNumber)
		if err != nil {
			if errors.Is(err, commons.ErrRecordNotFound) {
				return fmt.Errorf("credit account not found")
			}
			return fmt.Errorf("failed to fetch credit account")
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "debit account not found") {
			return commons.ErrorResponse[models.InternalTransferResponse]("Debit account not found"), err
		}
		if strings.Contains(errMsg, "credit account not found") {
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
	pinVerificationResp, pinVerificationErr := s.userService.VerifyUserPin(
		ctx,
		debitAccount.CustomerID,
		strings.TrimSpace(req.TransactionPIN),
	)
	if pinVerificationErr != nil {
		if pinVerificationResp.Message == "invalid pin" {
			err := fmt.Errorf("invalid transactionPIN")
			return commons.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error()), err
		}
		return commons.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), pinVerificationErr
	}
	if pinVerificationResp.Data == nil || !pinVerificationResp.Data.IsValidPin {
		err := fmt.Errorf("invalid transactionPIN")
		return commons.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error()), err
	}

	// Fetch rate conversion and charges in parallel
	var convertedAmount, rateUsed, chargeAmount, vatAmount, sumTotal decimal.Decimal
	eg2, eg2Ctx := errgroup.WithContext(ctx)

	eg2.Go(func() error {
		var err error
		var tempRateUsed, _ decimal.Decimal
		convertedAmount, tempRateUsed, _, err = s.rateService.ConvertRate(eg2Ctx, debitAmount, debitCurrency, creditCurrency)
		rateUsed = tempRateUsed
		return err
	})

	eg2.Go(func() error {
		var err error
		var tempChargeAmount, tempVatAmount decimal.Decimal
		_, _, tempChargeAmount, tempVatAmount, sumTotal, err = s.chargeService.GetCharges(eg2Ctx, debitAmount, debitCurrency)
		chargeAmount = tempChargeAmount
		vatAmount = tempVatAmount
		return err
	})

	if err := eg2.Wait(); err != nil {
		return commons.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
	}

	creditAmount := convertedAmount

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
			DebitBankName:        stringPtr(req.DebitBankName),
			CreditBankName:       stringPtr(req.CreditBankName),
			DebitCurrency:        debitCurrency,
			CreditCurrency:       creditCurrency,
			DebitAmount:          debitAmount,
			CreditAmount:         creditAmount,
			FCYRate:              rateUsed,
			ChargeAmount:         chargeAmount,
			VATAmount:            vatAmount,
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
		sumTotal,
		s.internalTransientAccountNumber,
		debitAmount,
		creditAccountNumber,
		creditAmount,
	)
	if postingErr != nil {
		_ = s.transferRepo.UpdateStatus(ctx, createdTransfer.ID, domain.TransferStatusFailed)
		if strings.Contains(strings.ToLower(postingErr.Error()), "insufficient balance") {
			err := commons.ErrInsufficientBalance
			return commons.ErrorResponse[models.InternalTransferResponse]("Insufficient balance", err.Error()), err
		}
		return commons.ErrorResponse[models.InternalTransferResponse]("transfer failed", "Unable to complete transfer posting"), postingErr
	}

	// Create both transient account transaction entries in parallel
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = s.transientAccountTransactionRepo.Create(ctx, domain.TransientAccountTransaction{
			TransferID:        createdTransfer.ID,
			ExternalRefernece: valueOrEmpty(createdTransfer.TransactionReference),
			DebitedAccount:    debitAccountNumber,
			CreditedAccount:   s.internalTransientAccountNumber,
			EntryType:         domain.LedgerEntryCredit,
			Currency:          debitCurrency,
			Amount:            sumTotal,
		})
	}()
	go func() {
		defer wg.Done()
		_, _ = s.transientAccountTransactionRepo.Create(ctx, domain.TransientAccountTransaction{
			TransferID:        createdTransfer.ID,
			ExternalRefernece: valueOrEmpty(createdTransfer.TransactionReference),
			DebitedAccount:    s.internalTransientAccountNumber,
			CreditedAccount:   creditAccountNumber,
			EntryType:         domain.LedgerEntryDebit,
			Currency:          creditCurrency,
			Amount:            creditAmount,
		})
	}()
	wg.Wait()

	_ = s.transferRepo.UpdateStatus(ctx, createdTransfer.ID, domain.TransferStatusSuccess)
	createdTransfer.Status = domain.TransferStatusSuccess

	chargeUSD, vatUSD, err := s.convertFeesToUSD(ctx, chargeAmount, vatAmount, debitCurrency)
	if err != nil {
		logger.Error("transfer service convert settlement fees to usd failed", err, logger.Fields{
			"transferId": createdTransfer.ID,
		})
		response := mapTransferToResponse(createdTransfer, sumTotal)
		return commons.SuccessResponse("Transaction successful. Settlement pending", response), nil
	}

	settlementErr := s.transientAccountRepo.SettleFromSuspenseToFees(
		ctx,
		s.internalTransientAccountNumber,
		chargeAmount,
		vatAmount,
		s.internalChargesAccountNumber,
		s.internalVATAccountNumber,
		chargeUSD,
		vatUSD,
	)
	if settlementErr != nil {
		logger.Error("transfer service settlement failed", settlementErr, logger.Fields{
			"transferId": createdTransfer.ID,
		})
		response := mapTransferToResponse(createdTransfer, sumTotal)
		return commons.SuccessResponse("Transaction successful. Settlement pending", response), nil
	}

	// Create all fee settlement transaction entries in parallel
	var wg2 sync.WaitGroup
	wg2.Add(4)
	go func() {
		defer wg2.Done()
		_, _ = s.transientAccountTransactionRepo.Create(ctx, domain.TransientAccountTransaction{
			TransferID:        createdTransfer.ID,
			ExternalRefernece: valueOrEmpty(createdTransfer.TransactionReference),
			DebitedAccount:    s.internalTransientAccountNumber,
			CreditedAccount:   s.internalChargesAccountNumber,
			EntryType:         domain.LedgerEntryDebit,
			Currency:          debitCurrency,
			Amount:            chargeAmount,
		})
	}()
	go func() {
		defer wg2.Done()
		_, _ = s.transientAccountTransactionRepo.Create(ctx, domain.TransientAccountTransaction{
			TransferID:        createdTransfer.ID,
			ExternalRefernece: valueOrEmpty(createdTransfer.TransactionReference),
			DebitedAccount:    s.internalTransientAccountNumber,
			CreditedAccount:   s.internalVATAccountNumber,
			EntryType:         domain.LedgerEntryDebit,
			Currency:          debitCurrency,
			Amount:            vatAmount,
		})
	}()
	go func() {
		defer wg2.Done()
		_, _ = s.transientAccountTransactionRepo.Create(ctx, domain.TransientAccountTransaction{
			TransferID:        createdTransfer.ID,
			ExternalRefernece: valueOrEmpty(createdTransfer.TransactionReference),
			DebitedAccount:    s.internalTransientAccountNumber,
			CreditedAccount:   s.internalChargesAccountNumber,
			EntryType:         domain.LedgerEntryCredit,
			Currency:          "USD",
			Amount:            chargeUSD,
		})
	}()
	go func() {
		defer wg2.Done()
		_, _ = s.transientAccountTransactionRepo.Create(ctx, domain.TransientAccountTransaction{
			TransferID:        createdTransfer.ID,
			ExternalRefernece: valueOrEmpty(createdTransfer.TransactionReference),
			DebitedAccount:    s.internalTransientAccountNumber,
			CreditedAccount:   s.internalVATAccountNumber,
			EntryType:         domain.LedgerEntryCredit,
			Currency:          "USD",
			Amount:            vatUSD,
		})
	}()
	wg2.Wait()

	_ = s.transferRepo.UpdateStatus(ctx, createdTransfer.ID, domain.TransferStatusClosed)
	createdTransfer.Status = domain.TransferStatusClosed

	response := mapTransferToResponse(createdTransfer, sumTotal)
	return commons.SuccessResponse("Transaction successful", response), nil
}

func (s *TransferService) processExternalTransfer(ctx context.Context, req models.InternalTransferRequest) (commons.Response[models.InternalTransferResponse], error) {
	beneficiaryBankCode := strings.TrimSpace(req.BeneficiaryBankCode)
	beneficiaryBankName, foundBankCode, err := s.getParticipantBankNameByCode(ctx, beneficiaryBankCode)
	if err != nil {
		return commons.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
	}
	if !foundBankCode {
		validationErr := fmt.Errorf("beneficiaryBankCode is not supported")
		return commons.ErrorResponse[models.InternalTransferResponse]("validation failed", validationErr.Error()), validationErr
	}

	debitAccountNumber := strings.TrimSpace(req.DebitAccountNumber)
	debitCurrency := strings.ToUpper(strings.TrimSpace(req.DebitCurrency))
	creditCurrency := strings.ToUpper(strings.TrimSpace(req.CreditCurrency))
	debitAmount := req.DebitAmount
	creditAccountNumber := strings.TrimSpace(req.CreditAccountNumber)

	debitAccount, err := s.accountRepo.GetByAccountNumber(ctx, debitAccountNumber)
	if err != nil {
		if errors.Is(err, commons.ErrRecordNotFound) {
			return commons.ErrorResponse[models.InternalTransferResponse]("Debit account not found"), err
		}
		return commons.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
	}
	if debitAccount.Status != domain.AccountStatusActive {
		validationErr := fmt.Errorf("debit account is not active")
		return commons.ErrorResponse[models.InternalTransferResponse]("validation failed", validationErr.Error()), validationErr
	}
	if !strings.EqualFold(strings.TrimSpace(debitAccount.Currency), debitCurrency) {
		validationErr := fmt.Errorf("debit currency does not match debit account currency")
		return commons.ErrorResponse[models.InternalTransferResponse]("validation failed", validationErr.Error()), validationErr
	}

	pinVerificationResp, pinVerificationErr := s.userService.VerifyUserPin(
		ctx,
		debitAccount.CustomerID,
		strings.TrimSpace(req.TransactionPIN),
	)
	if pinVerificationErr != nil {
		if pinVerificationResp.Message == "invalid pin" {
			validationErr := fmt.Errorf("invalid transactionPIN")
			return commons.ErrorResponse[models.InternalTransferResponse]("validation failed", validationErr.Error()), validationErr
		}
		return commons.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), pinVerificationErr
	}
	if pinVerificationResp.Data == nil || !pinVerificationResp.Data.IsValidPin {
		validationErr := fmt.Errorf("invalid transactionPIN")
		return commons.ErrorResponse[models.InternalTransferResponse]("validation failed", validationErr.Error()), validationErr
	}

	creditAmount, rateUsed, _, err := s.rateService.ConvertRate(ctx, debitAmount, debitCurrency, creditCurrency)
	if err != nil {
		return commons.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
	}

	_, _, chargeAmount, vatAmount, sumTotal, err := s.chargeService.GetCharges(ctx, debitAmount, debitCurrency)
	if err != nil {
		return commons.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
	}

	chargeUSD, vatUSD, err := s.convertFeesToUSD(ctx, chargeAmount, vatAmount, debitCurrency)
	if err != nil {
		return commons.ErrorResponse[models.InternalTransferResponse]("failed to process transfer", "Unable to process transfer right now"), err
	}

	externalAccountNumber, err := s.resolveExternalGLAccountNumber(creditCurrency)
	if err != nil {
		return commons.ErrorResponse[models.InternalTransferResponse]("validation failed", err.Error()), err
	}

	narration := strings.TrimSpace(req.Narration)
	auditPayloadBytes, _ := json.Marshal(logger.SanitizePayload(req))
	auditPayload := string(auditPayloadBytes)

	var createdTransfer domain.Transfer
	for attempt := 0; attempt < 5; attempt++ {
		transactionReference := generateThirtyDigitTransferReference()
		externalReference := generateExternalTransferReference()
		transferRecord := domain.Transfer{
			ExternalRefernece:    stringPtr(externalReference),
			TransactionReference: stringPtr(transactionReference),
			DebitAccountNumber:   debitAccountNumber,
			CreditAccountNumber:  stringPtr(creditAccountNumber),
			BeneficiaryBankCode:  stringPtr(beneficiaryBankCode),
			DebitBankName:        stringPtr(req.DebitBankName),
			CreditBankName:       stringPtr(beneficiaryBankName),
			DebitCurrency:        debitCurrency,
			CreditCurrency:       creditCurrency,
			DebitAmount:          debitAmount,
			CreditAmount:         creditAmount,
			FCYRate:              rateUsed,
			ChargeAmount:         chargeAmount,
			VATAmount:            vatAmount,
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

	postingErr := s.transferRepo.ProcessExternalTransfer(
		ctx,
		debitAccountNumber,
		sumTotal,
		s.internalTransientAccountNumber,
		debitAmount,
		externalAccountNumber,
		creditAmount,
		creditCurrency,
	)
	if postingErr != nil {
		_ = s.transferRepo.UpdateStatus(ctx, createdTransfer.ID, domain.TransferStatusFailed)
		if strings.Contains(strings.ToLower(postingErr.Error()), "insufficient balance") {
			insufficientErr := commons.ErrInsufficientBalance
			return commons.ErrorResponse[models.InternalTransferResponse]("Insufficient balance", insufficientErr.Error()), insufficientErr
		}
		return commons.ErrorResponse[models.InternalTransferResponse]("transfer failed", "Unable to complete transfer posting"), postingErr
	}

	_, _ = s.transientAccountTransactionRepo.Create(ctx, domain.TransientAccountTransaction{
		TransferID:        createdTransfer.ID,
		ExternalRefernece: valueOrEmpty(createdTransfer.ExternalRefernece),
		DebitedAccount:    debitAccountNumber,
		CreditedAccount:   s.internalTransientAccountNumber,
		EntryType:         domain.LedgerEntryCredit,
		Currency:          debitCurrency,
		Amount:            sumTotal,
	})
	_, _ = s.transientAccountTransactionRepo.Create(ctx, domain.TransientAccountTransaction{
		TransferID:        createdTransfer.ID,
		ExternalRefernece: valueOrEmpty(createdTransfer.ExternalRefernece),
		DebitedAccount:    s.internalTransientAccountNumber,
		CreditedAccount:   externalAccountNumber,
		EntryType:         domain.LedgerEntryDebit,
		Currency:          creditCurrency,
		Amount:            creditAmount,
	})

	_ = s.transferRepo.UpdateStatus(ctx, createdTransfer.ID, domain.TransferStatusSuccess)
	createdTransfer.Status = domain.TransferStatusSuccess

	settlementErr := s.transientAccountRepo.SettleFromSuspenseToFees(
		ctx,
		s.internalTransientAccountNumber,
		chargeAmount,
		vatAmount,
		s.internalChargesAccountNumber,
		s.internalVATAccountNumber,
		chargeUSD,
		vatUSD,
	)
	if settlementErr != nil {
		logger.Error("transfer service external settlement failed", settlementErr, logger.Fields{
			"transferId": createdTransfer.ID,
		})
		response := mapTransferToResponse(createdTransfer, sumTotal)
		return commons.SuccessResponse("Transaction successful. Settlement pending", response), nil
	}

	_, _ = s.transientAccountTransactionRepo.Create(ctx, domain.TransientAccountTransaction{
		TransferID:        createdTransfer.ID,
		ExternalRefernece: valueOrEmpty(createdTransfer.ExternalRefernece),
		DebitedAccount:    s.internalTransientAccountNumber,
		CreditedAccount:   s.internalChargesAccountNumber,
		EntryType:         domain.LedgerEntryDebit,
		Currency:          debitCurrency,
		Amount:            chargeAmount,
	})
	_, _ = s.transientAccountTransactionRepo.Create(ctx, domain.TransientAccountTransaction{
		TransferID:        createdTransfer.ID,
		ExternalRefernece: valueOrEmpty(createdTransfer.ExternalRefernece),
		DebitedAccount:    s.internalTransientAccountNumber,
		CreditedAccount:   s.internalVATAccountNumber,
		EntryType:         domain.LedgerEntryDebit,
		Currency:          debitCurrency,
		Amount:            vatAmount,
	})
	_, _ = s.transientAccountTransactionRepo.Create(ctx, domain.TransientAccountTransaction{
		TransferID:        createdTransfer.ID,
		ExternalRefernece: valueOrEmpty(createdTransfer.ExternalRefernece),
		DebitedAccount:    s.internalTransientAccountNumber,
		CreditedAccount:   s.internalChargesAccountNumber,
		EntryType:         domain.LedgerEntryCredit,
		Currency:          "USD",
		Amount:            chargeUSD,
	})
	_, _ = s.transientAccountTransactionRepo.Create(ctx, domain.TransientAccountTransaction{
		TransferID:        createdTransfer.ID,
		ExternalRefernece: valueOrEmpty(createdTransfer.ExternalRefernece),
		DebitedAccount:    s.internalTransientAccountNumber,
		CreditedAccount:   s.internalVATAccountNumber,
		EntryType:         domain.LedgerEntryCredit,
		Currency:          "USD",
		Amount:            vatUSD,
	})

	_ = s.transferRepo.UpdateStatus(ctx, createdTransfer.ID, domain.TransferStatusClosed)
	createdTransfer.Status = domain.TransferStatusClosed

	response := mapTransferToResponse(createdTransfer, sumTotal)
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

	rateValue := rateResp.Data.Rate
	if rateValue.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, decimal.Zero, fmt.Errorf("usd rate must be greater than zero")
	}

	chargeUSD := chargeAmount.Mul(rateValue)
	vatUSD := vatAmount.Mul(rateValue)

	return chargeUSD, vatUSD, nil
}

func mapTransferToResponse(transfer domain.Transfer, sumTotal decimal.Decimal) models.InternalTransferResponse {
	return models.InternalTransferResponse{
		TransactionReference: valueOrEmpty(transfer.TransactionReference),
		ExternalReference:    valueOrEmpty(transfer.ExternalRefernece),
		DebitAccountNumber:   transfer.DebitAccountNumber,
		CreditAccountNumber:  valueOrEmpty(transfer.CreditAccountNumber),
		BeneficiaryBankCode:  valueOrEmpty(transfer.BeneficiaryBankCode),
		DebitCurrency:        transfer.DebitCurrency,
		CreditCurrency:       transfer.CreditCurrency,
		DebitAmount:          decimalPtr(transfer.DebitAmount),
		CreditAmount:         decimalPtr(transfer.CreditAmount),
		FcyRate:              decimalPtr(transfer.FCYRate),
		ChargeAmount:         decimalPtr(transfer.ChargeAmount),
		VATAmount:            decimalPtr(transfer.VATAmount),
		SumTotalDebit:        decimalPtr(sumTotal),
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

func generateExternalTransferReference() string {
	base := generateThirtyDigitTransferReference()
	return "EXT" + base[:27]
}

func (s *TransferService) getParticipantBankNameByCode(ctx context.Context, bankCode string) (string, bool, error) {
	banks, err := s.participantBankRepo.GetAll(ctx)
	if err != nil {
		return "", false, err
	}

	trimmedCode := strings.TrimSpace(bankCode)
	for _, bank := range banks {
		if strings.TrimSpace(bank.BankCode) == trimmedCode {
			return strings.TrimSpace(bank.BankName), true, nil
		}
	}
	return "", false, nil
}

func (s *TransferService) resolveExternalGLAccountNumber(currency string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "USD":
		return s.externalUSDGLAccountNumber, nil
	case "GBP":
		return s.externalGBPGLAccountNumber, nil
	case "EUR":
		return s.externalEURGLAccountNumber, nil
	case "NGN":
		return s.externalNGNGLAccountNumber, nil
	default:
		return "", fmt.Errorf("unsupported credit currency")
	}
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

func decimalPtr(value decimal.Decimal) *decimal.Decimal {
	v := value
	return &v
}
