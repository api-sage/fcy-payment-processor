package usecase

import (
	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
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
		greyBankCode:                    greyBankCode,
		internalTransientAccountNumber:  internalTransientAccountNumber,
		internalChargesAccountNumber:    internalChargesAccountNumber,
		internalVATAccountNumber:        internalVATAccountNumber,
	}
}
