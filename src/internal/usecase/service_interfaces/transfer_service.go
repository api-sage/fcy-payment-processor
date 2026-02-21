package service_interfaces

import (
	"context"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/commons"
)

type TransferService interface {
	TransferFunds(ctx context.Context, req models.InternalTransferRequest) (commons.Response[models.InternalTransferResponse], error)
}
