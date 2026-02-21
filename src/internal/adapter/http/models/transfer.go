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
	DebitAccountNumber  string          `json:"debitAccountNumber"`
	CreditAccountNumber string          `json:"creditAccountNumber"`
	BeneficiaryBankCode string          `json:"beneficiaryBankCode"`
	TransactionPIN      string          `json:"transactionPIN"`
	DebitBankName       string          `json:"debitBankName"`
	CreditBankName      string          `json:"creditBankName"`
	DebitCurrency       string          `json:"debitCurrency"`
	CreditCurrency      string          `json:"creditCurrency"`
	DebitAmount         decimal.Decimal `json:"debitAmount"`
	Narration           string          `json:"narration"`
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
	if strings.TrimSpace(r.TransactionPIN) == "" {
		errs = append(errs, "transactionPIN is required")
	}
	if strings.TrimSpace(r.DebitBankName) == "" {
		errs = append(errs, "debitBankName is required")
	}
	if strings.TrimSpace(r.CreditBankName) == "" {
		errs = append(errs, "creditBankName is required")
	}

	debitCurrency := strings.ToUpper(strings.TrimSpace(r.DebitCurrency))
	creditCurrency := strings.ToUpper(strings.TrimSpace(r.CreditCurrency))
	if len(debitCurrency) != 3 {
		errs = append(errs, "debitCurrency must be 3 characters")
	}
	if len(creditCurrency) != 3 {
		errs = append(errs, "creditCurrency must be 3 characters")
	}

	if r.DebitAmount.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, "debitAmount must be greater than zero")
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
	TransactionReference string           `json:"transactionReference"`
	ExternalReference    string           `json:"externalReference"`
	DebitAccountNumber   string           `json:"debitAccountNumber"`
	CreditAccountNumber  string           `json:"creditAccountNumber"`
	BeneficiaryBankCode  string           `json:"beneficiaryBankCode"`
	DebitCurrency        string           `json:"debitCurrency"`
	CreditCurrency       string           `json:"creditCurrency"`
	DebitAmount          *decimal.Decimal `json:"debitAmount"`
	CreditAmount         *decimal.Decimal `json:"creditAmount"`
	FcyRate              *decimal.Decimal `json:"fcyRate"`
	ChargeAmount         *decimal.Decimal `json:"chargeAmount"`
	VATAmount            *decimal.Decimal `json:"vatAmount"`
	SumTotalDebit        *decimal.Decimal `json:"sumTotalDebit"`
	Narration            string           `json:"narration"`
	Status               string           `json:"status"`
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
