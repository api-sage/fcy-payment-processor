package service_interfaces

import (
	"context"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/commons"
)

type TransactionService interface {
	GetTransaction(ctx context.Context, req models.GetTransactionRequest) (commons.Response[models.GetTransactionResponse], error)
}
