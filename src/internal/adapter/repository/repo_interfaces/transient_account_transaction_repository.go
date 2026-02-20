package repo_interfaces

import (
	"context"

	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
)

type TransientAccountTransactionRepository interface {
	Create(ctx context.Context, entry domain.TransientAccountTransaction) (domain.TransientAccountTransaction, error)
}
