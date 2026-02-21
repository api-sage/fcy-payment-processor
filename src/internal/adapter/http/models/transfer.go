package models

import (
	"errors"
	"strings"

	"github.com/shopspring/decimal"
)

var allowedNarrations = []string{
	"Travels and Holiday",
	"Salary",
	"Project charge",
	"Food and consumables",
	"Transportation",
	"Accomodation",
	"utility bill",
	"savings",
	"investment",
	"loan",
	"loan repayment",
	"others",
}

type InternalTransferRequest struct {
	DebitAccountNumber  string `json:"debitAccountNumber"`
	CreditAccountNumber string `json:"creditAccountNumber"`
	BeneficiaryBankCode string `json:"beneficiaryBankCode"`
	DebitCurrency       string `json:"debitCurrency"`
	CreditCurrency      string `json:"creditCurrency"`
	DebitAmount         string `json:"debitAmount"`
	Narration           string `json:"narration"`
}

func (r InternalTransferRequest) Validate() error {
	var errs []string

	if !isTenDigits(r.DebitAccountNumber) {
		errs = append(errs, "debitAccountNumber must be exactly 10 digits")
	}
	if !isTenDigits(r.CreditAccountNumber) {
		errs = append(errs, "creditAccountNumber must be exactly 10 digits")
	}

	beneficiaryBankCode := strings.TrimSpace(r.BeneficiaryBankCode)
	if len(beneficiaryBankCode) != 6 || !digitsOnly(beneficiaryBankCode) {
		errs = append(errs, "beneficiaryBankCode must be exactly 6 digits")
	}

	debitCurrency := strings.ToUpper(strings.TrimSpace(r.DebitCurrency))
	creditCurrency := strings.ToUpper(strings.TrimSpace(r.CreditCurrency))
	if len(debitCurrency) != 3 {
		errs = append(errs, "debitCurrency must be 3 characters")
	}
	if len(creditCurrency) != 3 {
		errs = append(errs, "creditCurrency must be 3 characters")
	}

	amountRaw := strings.TrimSpace(r.DebitAmount)
	if amountRaw == "" {
		errs = append(errs, "debitAmount is required")
	} else {
		value, err := decimal.NewFromString(amountRaw)
		if err != nil {
			errs = append(errs, "debitAmount must be numeric")
		} else if value.LessThanOrEqual(decimal.Zero) {
			errs = append(errs, "debitAmount must be greater than zero")
		}
	}

	if !isAllowedNarration(strings.TrimSpace(r.Narration)) {
		errs = append(errs, "narration is not supported")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

type InternalTransferResponse struct {
	TransactionReference string `json:"transactionReference"`
	DebitAccountNumber   string `json:"debitAccountNumber"`
	CreditAccountNumber  string `json:"creditAccountNumber"`
	BeneficiaryBankCode  string `json:"beneficiaryBankCode"`
	DebitCurrency        string `json:"debitCurrency"`
	CreditCurrency       string `json:"creditCurrency"`
	DebitAmount          string `json:"debitAmount"`
	CreditAmount         string `json:"creditAmount"`
	FcyRate              string `json:"fcyRate"`
	ChargeAmount         string `json:"chargeAmount"`
	VATAmount            string `json:"vatAmount"`
	SumTotalDebit        string `json:"sumTotalDebit"`
	Narration            string `json:"narration"`
	Status               string `json:"status"`
}

func isAllowedNarration(value string) bool {
	for _, allowed := range allowedNarrations {
		if strings.EqualFold(strings.TrimSpace(allowed), value) {
			return true
		}
	}
	return false
}

func isTenDigits(value string) bool {
	trimmed := strings.TrimSpace(value)
	return len(trimmed) == 10 && digitsOnly(trimmed)
}

func digitsOnly(value string) bool {
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
