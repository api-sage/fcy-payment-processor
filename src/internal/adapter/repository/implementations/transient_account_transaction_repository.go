package implementations

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

func (r *TransientAccountTransactionRepository) List(ctx context.Context, filter domain.TransientAccountTransactionFilter) ([]domain.TransientAccountTransaction, int64, error) {
	logger.Info("transient account transaction repository list", logger.Fields{
		"startDate":       filter.StartDate,
		"endDate":         filter.EndDate,
		"glAccountNumber": filter.GLAccountNumber,
		"limit":           filter.Limit,
		"offset":          filter.Offset,
	})

	var (
		conditions []string
		args       []any
	)
	addArg := func(value any) string {
		args = append(args, value)
		return fmt.Sprintf("$%d", len(args))
	}

	if filter.StartDate != nil {
		conditions = append(conditions, "created_at >= "+addArg(*filter.StartDate))
	}
	if filter.EndDate != nil {
		conditions = append(conditions, "created_at < "+addArg(*filter.EndDate))
	}
	if filter.GLAccountNumber != "" {
		placeholder := addArg(filter.GLAccountNumber)
		conditions = append(conditions, fmt.Sprintf("(debited_account = %s OR credited_account = %s)", placeholder, placeholder))
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM transient_account_transactions " + whereClause
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		logger.Error("transient account transaction repository count failed", err, nil)
		return nil, 0, fmt.Errorf("count transient account transactions: %w", err)
	}

	listQuery := fmt.Sprintf(`
SELECT id, transfer_id, external_refernece, debited_account, credited_account, entry_type, currency, amount, created_at
FROM transient_account_transactions %s
ORDER BY created_at DESC, id DESC
LIMIT %s OFFSET %s`,
		whereClause,
		addArg(filter.Limit),
		addArg(filter.Offset),
	)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		logger.Error("transient account transaction repository list failed", err, nil)
		return nil, 0, fmt.Errorf("list transient account transactions: %w", err)
	}

	entries, err := scanTransientAccountTransactions(rows)
	if err != nil {
		return nil, 0, err
	}

	logger.Info("transient account transaction repository list success", logger.Fields{
		"count": len(entries),
		"total": total,
	})

	return entries, total, nil
}

// FindByExternalReference returns every transient account transaction that shares the given external reference, oldest first.
func (r *TransientAccountTransactionRepository) FindByExternalReference(ctx context.Context, reference string) ([]domain.TransientAccountTransaction, error) {
	logger.Info("transient account transaction repository find by external reference", logger.Fields{
		"externalReference": reference,
	})

	rows, err := r.db.QueryContext(ctx, `
SELECT id, transfer_id, external_refernece, debited_account, credited_account, entry_type, currency, amount, created_at
FROM transient_account_transactions
WHERE external_refernece = $1
ORDER BY created_at ASC, id ASC`, reference)
	if err != nil {
		logger.Error("transient account transaction repository find by external reference failed", err, nil)
		return nil, fmt.Errorf("find transient account transactions by external reference: %w", err)
	}

	entries, err := scanTransientAccountTransactions(rows)
	if err != nil {
		return nil, err
	}

	logger.Info("transient account transaction repository find by external reference success", logger.Fields{
		"count": len(entries),
	})

	return entries, nil
}

// scanTransientAccountTransactions reads and closes rows from a transient_account_transactions select.
func scanTransientAccountTransactions(rows *sql.Rows) ([]domain.TransientAccountTransaction, error) {
	defer rows.Close()

	entries := make([]domain.TransientAccountTransaction, 0)
	for rows.Next() {
		var (
			entry           domain.TransientAccountTransaction
			debitedAccount  sql.NullString
			creditedAccount sql.NullString
		)
		if err := rows.Scan(
			&entry.ID,
			&entry.TransferID,
			&entry.ExternalRefernece,
			&debitedAccount,
			&creditedAccount,
			&entry.EntryType,
			&entry.Currency,
			&entry.Amount,
			&entry.CreatedAt,
		); err != nil {
			logger.Error("transient account transaction repository scan failed", err, nil)
			return nil, fmt.Errorf("scan transient account transaction: %w", err)
		}
		entry.DebitedAccount = debitedAccount.String
		entry.CreditedAccount = creditedAccount.String
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		logger.Error("transient account transaction repository iterate failed", err, nil)
		return nil, fmt.Errorf("iterate transient account transactions: %w", err)
	}

	return entries, nil
}
