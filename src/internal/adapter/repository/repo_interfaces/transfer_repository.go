package repo_interfaces

import (
	"context"

	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
	"github.com/shopspring/decimal"
)

type TransferRepository interface {
	Create(ctx context.Context, transfer domain.Transfer) (domain.Transfer, error)
	Update(ctx context.Context, transfer domain.Transfer) (domain.Transfer, error)
	Get(ctx context.Context, id string, transactionReference string, externalRefernece string) (domain.Transfer, error)
	ProcessInternalTransfer(ctx context.Context, debitAccountNumber string, debitAmount decimal.Decimal, suspenseAccountNumber string, debitSuspenseAccountAmount decimal.Decimal, creditAccountNumber string, creditAmount decimal.Decimal) error
	ProcessExternalTransfer(
		ctx context.Context,
		debitAccountNumber string,
		totalDebitAmount decimal.Decimal,
		suspenseAccountNumber string,
		beneficiaryAmount decimal.Decimal,
		externalAccountNumber string,
		externalAccountCurrency string,
	) error
	UpdateStatus(ctx context.Context, transferID string, status domain.TransferStatus) error
}
