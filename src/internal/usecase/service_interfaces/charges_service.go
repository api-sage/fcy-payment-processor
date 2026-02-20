package service_interfaces

import (
	"context"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/commons"
)

type ChargesService interface {
	GetChargesSummary(ctx context.Context, req models.GetChargesRequest) (commons.Response[models.GetChargesResponse], error)
	GetCharges(amount string, fromCurrency string) (string, string, string, string, string, error)
}
