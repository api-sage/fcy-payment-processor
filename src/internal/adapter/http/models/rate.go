package models

import (
	"errors"
	"strings"

	"github.com/shopspring/decimal"
)

type RateResponse struct {
	ID           int64  `json:"id"`
	FromCurrency string `json:"fromCurrency"`
	ToCurrency   string `json:"toCurrency"`
	Rate         string `json:"rate"`
	RateDate     string `json:"rateDate"`
	CreatedAt    string `json:"createdAt"`
}

type GetRateRequest struct {
	FromCurrency string `json:"fromCurrency"`
	ToCurrency   string `json:"toCurrency"`
}

func (r GetRateRequest) Validate() error {
	var errs []string

	fromCurrency := strings.ToUpper(strings.TrimSpace(r.FromCurrency))
	toCurrency := strings.ToUpper(strings.TrimSpace(r.ToCurrency))

	if fromCurrency == "" {
		errs = append(errs, "fromCurrency is required")
	}
	if toCurrency == "" {
		errs = append(errs, "toCurrency is required")
	}
	if fromCurrency != "" && len(fromCurrency) != 3 {
		errs = append(errs, "fromCurrency must be 3 characters")
	}
	if toCurrency != "" && len(toCurrency) != 3 {
		errs = append(errs, "toCurrency must be 3 characters")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

type GetCcyRatesRequest struct {
	Amount  string `json:"amount"`
	FromCcy string `json:"fromCcy"`
	ToCcy   string `json:"toCcy"`
}

func (r GetCcyRatesRequest) Validate() error {
	var errs []string

	amount := strings.TrimSpace(r.Amount)
	fromCcy := strings.ToUpper(strings.TrimSpace(r.FromCcy))
	toCcy := strings.ToUpper(strings.TrimSpace(r.ToCcy))

	if amount == "" {
		errs = append(errs, "amount is required")
	} else {
		parsedAmount, err := decimal.NewFromString(amount)
		if err != nil {
			errs = append(errs, "amount must be numeric")
		} else if parsedAmount.LessThanOrEqual(decimal.Zero) {
			errs = append(errs, "amount must be greater than zero")
		}
	}

	if fromCcy == "" {
		errs = append(errs, "fromCcy is required")
	} else if len(fromCcy) != 3 {
		errs = append(errs, "fromCcy must be 3 characters")
	}

	if toCcy == "" {
		errs = append(errs, "toCcy is required")
	} else if len(toCcy) != 3 {
		errs = append(errs, "toCcy must be 3 characters")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

type GetCcyRatesResponse struct {
	Amount          string `json:"amount"`
	FromCcy         string `json:"fromCcy"`
	ToCcy           string `json:"toCcy"`
	ConvertedAmount string `json:"convertedAmount"`
	RateUsed        string `json:"rateUsed"`
	RateDate        string `json:"rateDate"`
}
