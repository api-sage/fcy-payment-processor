package models

import (
	"errors"
	"strings"

	"github.com/shopspring/decimal"
)

type GetChargesRequest struct {
	Amount       decimal.Decimal `json:"amount"`
	FromCurrency string          `json:"fromCurrency"`
}

func (r GetChargesRequest) Validate() error {
	var errs []string

	ccy := strings.ToUpper(strings.TrimSpace(r.FromCurrency))

	if r.Amount.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, "amount must be greater than zero")
	}

	if ccy == "" {
		errs = append(errs, "fromCurrency is required")
	} else if len(ccy) != 3 {
		errs = append(errs, "fromCurrency must be 3 characters")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

type GetChargesResponse struct {
	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`
	Charge   decimal.Decimal `json:"charge"`
	VAT      decimal.Decimal `json:"vat"`
	SumTotal decimal.Decimal `json:"sumTotal"`
}
