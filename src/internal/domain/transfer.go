package domain

import "time"

type TransferStatus string

const (
	TransferStatusPending TransferStatus = "PENDING"
	TransferStatusSuccess TransferStatus = "SUCCESS"
	TransferStatusFailed  TransferStatus = "FAILED"
)

type Transfer struct {
	ID                  string
	PaymentReference    *string
	DebitAccountNumber  string
	CreditAccountNumber *string
	BeneficiaryBankCode *string
	DebitCurrency       string
	CreditCurrency      string
	DebitAmount         string
	CreditAmount        string
	CcyRate             string
	ChargeAmount        string
	VATAmount           string
	Status              TransferStatus
	AuditPayload        string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	ProcessedAt         *time.Time
}
