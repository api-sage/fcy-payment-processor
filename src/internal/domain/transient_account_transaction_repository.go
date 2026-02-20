package domain

import "context"

type TransientAccountTransactionRepository interface {
	Create(ctx context.Context, entry TransientAccountTransaction) (TransientAccountTransaction, error)
}
