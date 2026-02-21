package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type LedgerEntryType string

const (
	LedgerEntryDebit  LedgerEntryType = "DEBIT"
	LedgerEntryCredit LedgerEntryType = "CREDIT"
)

type TransientAccountTransaction struct {
	ID                string
	TransferID        string
	ExternalRefernece string
	DebitedAccount    string
	CreditedAccount   string
	EntryType         LedgerEntryType
	Currency          string
	Amount            decimal.Decimal
	CreatedAt         time.Time
}
