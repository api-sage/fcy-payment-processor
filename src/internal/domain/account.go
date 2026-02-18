package domain

import "time"

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
	AvailableBalance string
	LedgerBalance    string
	Status           AccountStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
