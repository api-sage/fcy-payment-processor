package repo_interfaces

import (
	"context"

	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
)

type TransferRepository interface {
	Create(ctx context.Context, transfer domain.Transfer) (domain.Transfer, error)
	Update(ctx context.Context, transfer domain.Transfer) (domain.Transfer, error)
	Get(ctx context.Context, id string, transactionReference string, externalRefernece string) (domain.Transfer, error)
	ProcessInternalTransfer(ctx context.Context, debitAccountNumber string, debitAmount string, suspenseAccountNumber string, debitSuspenseAccountAmount string, creditAccountNumber string, creditAmount string) error
	UpdateStatus(ctx context.Context, transferID string, status domain.TransferStatus) error
}
