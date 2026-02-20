package service_interfaces

import (
	"context"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/commons"
)

type RateService interface {
	GetRates(ctx context.Context) (commons.Response[[]models.RateResponse], error)
	GetRate(ctx context.Context, req models.GetRateRequest) (commons.Response[models.RateResponse], error)
	ConvertRate(ctx context.Context, amount string, fromCcy string, toCcy string) (string, string, string, error)
	GetCcyRates(ctx context.Context, req models.GetCcyRatesRequest) (commons.Response[models.GetCcyRatesResponse], error)
}
