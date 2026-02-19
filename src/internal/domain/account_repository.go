package domain

import "context"

type AccountRepository interface {
	Create(ctx context.Context, account Account) (Account, error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (Account, error)
}
