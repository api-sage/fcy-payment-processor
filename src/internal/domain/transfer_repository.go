package domain

import "context"

type TransferRepository interface {
	Create(ctx context.Context, transfer Transfer) (Transfer, error)
	Update(ctx context.Context, transfer Transfer) (Transfer, error)
	Get(ctx context.Context, id string, transactionReference string, externalRefernece string) (Transfer, error)
	ProcessInternalTransfer(ctx context.Context, debitAccountNumber string, debitAmount string, suspenseAccountNumber string, creditAccountNumber string, creditAmount string) error
	UpdateStatus(ctx context.Context, transferID string, status TransferStatus) error
}
