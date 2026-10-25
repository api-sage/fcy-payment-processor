package service_interfaces

import (
	"context"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/commons"
)

type GLTransactionService interface {
	GetGLTransactions(ctx context.Context, req models.GetGLTransactionsRequest) (commons.Response[models.GetGLTransactionsResponse], error)
}
