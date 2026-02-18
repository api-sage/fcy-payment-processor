package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account domain.Account) (domain.Account, error) {
	const query = `
INSERT INTO accounts (
	customer_id,
	account_number,
	currency,
	available_balance,
	ledger_balance,
	status
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at, updated_at`

	var createdAt time.Time
	var updatedAt time.Time
	var id string

	if err := r.db.QueryRowContext(
		ctx,
		query,
		account.CustomerID,
		account.AccountNumber,
		account.Currency,
		account.AvailableBalance,
		account.LedgerBalance,
		account.Status,
	).Scan(&id, &createdAt, &updatedAt); err != nil {
		return domain.Account{}, fmt.Errorf("create account: %w", err)
	}

	account.ID = id
	account.CreatedAt = createdAt
	account.UpdatedAt = updatedAt

	return account, nil
}
