package domain

import "time"

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
	DebitCurrency        string
	CreditCurrency       string
	DebitAmount          string
	CreditAmount         string
	FCYRate              string
	ChargeAmount         string
	VATAmount            string
	Narration            *string
	Status               TransferStatus
	AuditPayload         string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ProcessedAt          *time.Time
}
