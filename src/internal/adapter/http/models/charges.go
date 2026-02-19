package models

import (
	"errors"
	"strconv"
	"strings"
)

type GetChargesRequest struct {
	Amount       string `json:"amount"`
	FromCurrency string `json:"fromCurrency"`
}

func (r GetChargesRequest) Validate() error {
	var errs []string

	amount := strings.TrimSpace(r.Amount)
	ccy := strings.ToUpper(strings.TrimSpace(r.FromCurrency))

	if amount == "" {
		errs = append(errs, "amount is required")
	} else {
		parsed, err := strconv.ParseFloat(amount, 64)
		if err != nil {
			errs = append(errs, "amount must be numeric")
		} else if parsed <= 0 {
			errs = append(errs, "amount must be greater than zero")
		}
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
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	Charge   string `json:"charge"`
	VAT      string `json:"vat"`
	SumTotal string `json:"sumTotal"`
}
