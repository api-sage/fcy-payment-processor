package models

import (
	"errors"
	"strings"
)

type RateResponse struct {
	ID           int64  `json:"id"`
	FromCurrency string `json:"fromCurrency"`
	ToCurrency   string `json:"toCurrency"`
	SellRate     string `json:"sellRate"`
	BuyRate      string `json:"buyRate"`
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
	if fromCurrency != "" && toCurrency != "" && fromCurrency == toCurrency {
		errs = append(errs, "fromCurrency and toCurrency cannot be the same")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}
