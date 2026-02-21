package service_interfaces

import (
	"context"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/commons"
	"github.com/shopspring/decimal"
)

type RateService interface {
	GetRates(ctx context.Context) (commons.Response[[]models.RateResponse], error)
	GetRate(ctx context.Context, req models.GetRateRequest) (commons.Response[models.RateResponse], error)
	ConvertRate(ctx context.Context, amount decimal.Decimal, fromCcy string, toCcy string) (decimal.Decimal, decimal.Decimal, string, error)
	GetCcyRates(ctx context.Context, req models.GetCcyRatesRequest) (commons.Response[models.GetCcyRatesResponse], error)
}
