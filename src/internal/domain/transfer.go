package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type TransferStatus string

const (
	TransferStatusPending TransferStatus = "PENDING"
	TransferStatusSuccess TransferStatus = "SUCCESS"
	TransferStatusFailed  TransferStatus = "FAILED"
	TransferStatusClosed  TransferStatus = "CLOSED"
)

type Transfer struct {
	ID                   string
	ExternalRefernece    *string
	TransactionReference *string
	DebitAccountNumber   string
	CreditAccountNumber  *string
	BeneficiaryBankCode  *string
	DebitBankName        *string
	CreditBankName       *string
	DebitCurrency        string
	CreditCurrency       string
	DebitAmount          decimal.Decimal
	CreditAmount         decimal.Decimal
	FCYRate              decimal.Decimal
	ChargeAmount         decimal.Decimal
	VATAmount            decimal.Decimal
	Narration            *string
	Status               TransferStatus
	AuditPayload         string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ProcessedAt          *time.Time
}
