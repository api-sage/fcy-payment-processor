package domain

import "context"

type AccountRepository interface {
	Create(ctx context.Context, account Account) (Account, error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (Account, error)
	DebitInternalAccount(ctx context.Context, accountNumber string, amount string) error
	CreditInternalAccount(ctx context.Context, accountNumber string, amount string) error
	DepositFunds(ctx context.Context, accountNumber string, amount string) error
}
