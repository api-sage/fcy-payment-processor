package interfaces

import (
	"context"

	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
)

type RateRepository interface {
	GetRates(ctx context.Context) ([]domain.Rate, error)
	GetRate(ctx context.Context, fromCurrency string, toCurrency string) (domain.Rate, error)
}
