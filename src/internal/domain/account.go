package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type AccountStatus string

const (
	AccountStatusActive AccountStatus = "ACTIVE"
	AccountStatusFrozen AccountStatus = "FROZEN"
	AccountStatusClosed AccountStatus = "CLOSED"
)

type Account struct {
	ID               string
	CustomerID       string
	AccountNumber    string
	Currency         string
	AvailableBalance decimal.Decimal
	LedgerBalance    decimal.Decimal
	Status           AccountStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
