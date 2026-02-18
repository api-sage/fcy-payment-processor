package domain

import "time"

type LedgerEntryType string

const (
	LedgerEntryDebit  LedgerEntryType = "DEBIT"
	LedgerEntryCredit LedgerEntryType = "CREDIT"
)

type TransientAccountTransaction struct {
	ID               string
	TransferID       string
	PaymentReference string
	EntryType        LedgerEntryType
	Currency         string
	Amount           string
	CreatedAt        time.Time
}
