package repo_interfaces

import (
	"context"

	"github.com/api-sage/fcy-payment-processor/src/internal/domain"
)

type TransientAccountTransactionRepository interface {
	Create(ctx context.Context, entry domain.TransientAccountTransaction) (domain.TransientAccountTransaction, error)
}
