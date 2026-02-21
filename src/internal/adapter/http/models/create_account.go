package models

import (
	"errors"
	"strings"

	"github.com/shopspring/decimal"
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
	} else if ccy != "USD" && ccy != "EUR" && ccy != "GBP" && ccy != "NGN" {
		errs = append(errs, "currency must be one of USD, EUR, GBP, NGN")
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

type DepositFundsRequest struct {
	AccountNumber string `json:"accountNumber"`
	Amount        string `json:"amount"`
}

func (r DepositFundsRequest) Validate() error {
	var errs []string

	accountNumber := strings.TrimSpace(r.AccountNumber)
	if len(accountNumber) != 10 {
		errs = append(errs, "accountNumber must be exactly 10 digits")
	} else {
		for _, ch := range accountNumber {
			if ch < '0' || ch > '9' {
				errs = append(errs, "accountNumber must be exactly 10 digits")
				break
			}
		}
	}

	amount := strings.TrimSpace(r.Amount)
	if amount == "" {
		errs = append(errs, "amount is required")
	} else {
		parsed, err := decimal.NewFromString(amount)
		if err != nil {
			errs = append(errs, "amount must be numeric")
		} else if parsed.LessThanOrEqual(decimal.Zero) {
			errs = append(errs, "amount must be greater than zero")
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

type DepositFundsResponse struct {
	AccountNumber    string `json:"accountNumber"`
	Currency         string `json:"currency"`
	DepositedAmount  string `json:"depositedAmount"`
	AvailableBalance string `json:"availableBalance"`
	LedgerBalance    string `json:"ledgerBalance"`
}
