package repo_interfaces

import (
	"context"

	"github.com/api-sage/fcy-payment-processor/src/internal/domain"
	"github.com/shopspring/decimal"
)

type AccountRepository interface {
	Create(ctx context.Context, account domain.Account) (domain.Account, error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (domain.Account, error)
	HasAccountForCustomerIDAndCurrency(ctx context.Context, customerID string, currency string) (bool, error)
	DebitInternalAccount(ctx context.Context, accountNumber string, amount decimal.Decimal) error
	CreditInternalAccount(ctx context.Context, accountNumber string, amount decimal.Decimal) error
	DepositFunds(ctx context.Context, accountNumber string, amount decimal.Decimal) error
}
