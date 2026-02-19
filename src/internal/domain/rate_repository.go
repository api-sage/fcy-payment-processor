package domain

import "context"

type RateRepository interface {
	GetRates(ctx context.Context) ([]Rate, error)
	GetRate(ctx context.Context, fromCurrency string, toCurrency string) (Rate, error)
}
