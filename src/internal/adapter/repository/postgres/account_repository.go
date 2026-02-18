package postgres

import (
	"context"
	"database/sql"
	"errors"
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

func (r *AccountRepository) GetByAccountNumber(ctx context.Context, accountNumber string) (domain.Account, error) {
	const query = `
SELECT id, customer_id, account_number, currency, available_balance, ledger_balance, status, created_at, updated_at
FROM accounts
WHERE account_number = $1`

	var account domain.Account
	if err := r.db.QueryRowContext(ctx, query, accountNumber).Scan(
		&account.ID,
		&account.CustomerID,
		&account.AccountNumber,
		&account.Currency,
		&account.AvailableBalance,
		&account.LedgerBalance,
		&account.Status,
		&account.CreatedAt,
		&account.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Account{}, fmt.Errorf("account not found: %w", err)
		}
		return domain.Account{}, fmt.Errorf("get account by account number: %w", err)
	}

	return account, nil
}
