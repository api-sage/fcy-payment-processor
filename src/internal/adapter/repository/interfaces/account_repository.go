package interfaces

import (
	"context"

	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
)

type AccountRepository interface {
	Create(ctx context.Context, account domain.Account) (domain.Account, error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (domain.Account, error)
	DebitInternalAccount(ctx context.Context, accountNumber string, amount string) error
	CreditInternalAccount(ctx context.Context, accountNumber string, amount string) error
	DepositFunds(ctx context.Context, accountNumber string, amount string) error
}
