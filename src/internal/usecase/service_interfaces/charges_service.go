package service_interfaces

import (
	"context"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/commons"
	"github.com/shopspring/decimal"
)

type ChargesService interface {
	GetChargesSummary(ctx context.Context, req models.GetChargesRequest) (commons.Response[models.GetChargesResponse], error)
	GetCharges(ctx context.Context, amount decimal.Decimal, fromCurrency string) (decimal.Decimal, string, decimal.Decimal, decimal.Decimal, decimal.Decimal, error)
}
