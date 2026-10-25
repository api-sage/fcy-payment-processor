package repo_interfaces

import (
	"context"

	"github.com/api-sage/fcy-payment-processor/src/internal/domain"
)

type TransientAccountTransactionRepository interface {
	Create(ctx context.Context, entry domain.TransientAccountTransaction) (domain.TransientAccountTransaction, error)
	List(ctx context.Context, filter domain.TransientAccountTransactionFilter) ([]domain.TransientAccountTransaction, int64, error)
	FindByExternalReference(ctx context.Context, reference string) ([]domain.TransientAccountTransaction, error)
}
