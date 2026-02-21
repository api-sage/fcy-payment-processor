package implementations

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/domain"
	"github.com/api-sage/fcy-payment-processor/src/internal/logger"
)

type TransientAccountTransactionRepository struct {
	db *sql.DB
}

func NewTransientAccountTransactionRepository(db *sql.DB) *TransientAccountTransactionRepository {
	return &TransientAccountTransactionRepository{db: db}
}

func (r *TransientAccountTransactionRepository) Create(ctx context.Context, entry domain.TransientAccountTransaction) (domain.TransientAccountTransaction, error) {
	logger.Info("transient account transaction repository create", logger.Fields{
		"transferId":        entry.TransferID,
		"externalRefernece": entry.ExternalRefernece,
		"debitedAccount":    entry.DebitedAccount,
		"creditedAccount":   entry.CreditedAccount,
		"entryType":         entry.EntryType,
		"currency":          entry.Currency,
		"amount":            entry.Amount,
	})

	const query = `
INSERT INTO transient_account_transactions (
	transfer_id,
	external_refernece,
	debited_account,
	credited_account,
	entry_type,
	currency,
	amount
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at`

	var (
		id        string
		createdAt time.Time
	)

	if err := r.db.QueryRowContext(
		ctx,
		query,
		entry.TransferID,
		entry.ExternalRefernece,
		entry.DebitedAccount,
		entry.CreditedAccount,
		entry.EntryType,
		entry.Currency,
		entry.Amount,
	).Scan(&id, &createdAt); err != nil {
		logger.Error("transient account transaction repository create failed", err, logger.Fields{
			"transferId": entry.TransferID,
		})
		return domain.TransientAccountTransaction{}, fmt.Errorf("create transient account transaction: %w", err)
	}

	entry.ID = id
	entry.CreatedAt = createdAt

	logger.Info("transient account transaction repository create success", logger.Fields{
		"id":         entry.ID,
		"transferId": entry.TransferID,
	})

	return entry, nil
}
