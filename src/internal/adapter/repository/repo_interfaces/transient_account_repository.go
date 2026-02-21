package repo_interfaces

import (
	"context"

	"github.com/shopspring/decimal"
)

type TransientAccountRepository interface {
	DebitSuspenseAccount(ctx context.Context, suspenseAccountNumber string, currency string, amount decimal.Decimal) error
	CreditSuspenseAccount(ctx context.Context, suspenseAccountNumber string, currency string, amount decimal.Decimal) error
	SettleFromSuspenseToFees(
		ctx context.Context,
		suspenseAccountNumber string,
		chargeAmount decimal.Decimal,
		vatAmount decimal.Decimal,
		chargesAccountNumber string,
		vatAccountNumber string,
		chargeUSD decimal.Decimal,
		vatUSD decimal.Decimal,
	) error
}
