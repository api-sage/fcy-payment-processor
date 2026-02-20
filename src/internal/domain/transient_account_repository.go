package domain

import "context"

type TransientAccountRepository interface {
	DebitSuspenseAccount(ctx context.Context, suspenseAccountNumber string, currency string, amount string) error
	CreditSuspenseAccount(ctx context.Context, suspenseAccountNumber string, currency string, amount string) error
	SettleFromSuspenseToFees(
		ctx context.Context,
		suspenseAccountNumber string,
		chargeAmount string,
		vatAmount string,
		chargesAccountNumber string,
		vatAccountNumber string,
		chargeUSD string,
		vatUSD string,
	) error
}
