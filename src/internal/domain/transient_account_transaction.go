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

// TransientAccountTransactionFilter narrows a transaction listing.
// StartDate is inclusive and EndDate is exclusive. A nil bound or an empty GLAccountNumber is not applied.
// GLAccountNumber matches either the debited or the credited account.
type TransientAccountTransactionFilter struct {
	StartDate       *time.Time
	EndDate         *time.Time
	GLAccountNumber string
	Limit           int
	Offset          int
}
