package domain

import "time"

type TransferStatus string

const (
	TransferStatusPending TransferStatus = "PENDING"
	TransferStatusSuccess TransferStatus = "SUCCESS"
	TransferStatusFailed  TransferStatus = "FAILED"
)

type Transfer struct {
	ID                   string
	PaymentReference     *string
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
	Status               TransferStatus
	AuditPayload         string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ProcessedAt          *time.Time
}
