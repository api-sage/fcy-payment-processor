package models

import (
	"errors"
	"strings"
)

type CreateAccountRequest struct {
	CustomerID     string `json:"customerId"`
	Currency       string `json:"currency"`
	InitialDeposit string `json:"initialDeposit,omitempty"`
}

func (r CreateAccountRequest) Validate() error {
	var errs []string

	if strings.TrimSpace(r.CustomerID) == "" {
		errs = append(errs, "customerId is required")
	}

	ccy := strings.ToUpper(strings.TrimSpace(r.Currency))
	if ccy == "" {
		errs = append(errs, "currency is required")
	} else if ccy != "USD" && ccy != "EUR" && ccy != "GBP" {
		errs = append(errs, "currency must be one of USD, EUR, GBP")
	}

	if strings.TrimSpace(r.InitialDeposit) != "" {
		if strings.HasPrefix(strings.TrimSpace(r.InitialDeposit), "-") {
			errs = append(errs, "initialDeposit cannot be negative")
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

type CreateAccountResponse struct {
	ID               string `json:"id"`
	CustomerID       string `json:"customerId"`
	AccountNumber    string `json:"accountNumber"`
	Currency         string `json:"currency"`
	AvailableBalance string `json:"availableBalance"`
	LedgerBalance    string `json:"ledgerBalance"`
	Status           string `json:"status"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

type GetAccountResponse struct {
	ID               string `json:"id"`
	CustomerID       string `json:"customerId"`
	AccountName      string `json:"accountName,omitempty"`
	AccountNumber    string `json:"accountNumber"`
	BankCode         string `json:"bankCode"`
	BankName         string `json:"bankName,omitempty"`
	Currency         string `json:"currency"`
	AvailableBalance string `json:"availableBalance"`
	LedgerBalance    string `json:"ledgerBalance"`
	Status           string `json:"status"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}
