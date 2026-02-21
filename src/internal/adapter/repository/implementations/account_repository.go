package implementations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/api-sage/ccy-payment-processor/src/internal/commons"
	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
	"github.com/api-sage/ccy-payment-processor/src/internal/logger"
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account domain.Account) (domain.Account, error) {
	logger.Info("account repository create", logger.Fields{
		"customerId":    account.CustomerID,
		"accountNumber": account.AccountNumber,
		"currency":      account.Currency,
	})

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
		logger.Error("account repository create failed", err, logger.Fields{
			"customerId":    account.CustomerID,
			"accountNumber": account.AccountNumber,
		})
		return domain.Account{}, fmt.Errorf("create account: %w", err)
	}

	account.ID = id
	account.CreatedAt = createdAt
	account.UpdatedAt = updatedAt
	logger.Info("account repository create success", logger.Fields{
		"accountId":     account.ID,
		"accountNumber": account.AccountNumber,
	})

	return account, nil
}

func (r *AccountRepository) GetByAccountNumber(ctx context.Context, accountNumber string) (domain.Account, error) {
	logger.Info("account repository get by account number", logger.Fields{
		"accountNumber": accountNumber,
	})

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
			logger.Info("account repository record not found", logger.Fields{
				"accountNumber": accountNumber,
			})
			return domain.Account{}, commons.ErrRecordNotFound
		}
		logger.Error("account repository get failed", err, logger.Fields{
			"accountNumber": accountNumber,
		})
		return domain.Account{}, fmt.Errorf("get account by account number: %w", err)
	}

	logger.Info("account repository get success", logger.Fields{
		"accountId":     account.ID,
		"accountNumber": account.AccountNumber,
	})

	return account, nil
}

func (r *AccountRepository) HasAccountForCustomerIDAndCurrency(ctx context.Context, customerID string, currency string) (bool, error) {
	logger.Info("account repository has account for customer id and currency", logger.Fields{
		"customerId": customerID,
		"currency":   currency,
	})

	const query = `
SELECT EXISTS (
	SELECT 1
	FROM accounts
	WHERE customer_id = $1
	  AND UPPER(currency) = UPPER($2)
)`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, customerID, currency).Scan(&exists); err != nil {
		logger.Error("account repository has account for customer id and currency failed", err, logger.Fields{
			"customerId": customerID,
			"currency":   currency,
		})
		return false, fmt.Errorf("check account by customer id and currency: %w", err)
	}

	return exists, nil
}

func (r *AccountRepository) DebitInternalAccount(ctx context.Context, accountNumber string, amount string) error {
	logger.Info("account repository debit internal account", logger.Fields{
		"accountNumber": accountNumber,
		"amount":        amount,
	})

	const query = `
UPDATE accounts
SET available_balance = available_balance - $2::numeric,
    updated_at = NOW()
WHERE account_number = $1
  AND status = 'ACTIVE'
  AND available_balance >= $2::numeric`

	result, err := r.db.ExecContext(ctx, query, accountNumber, amount)
	if err != nil {
		logger.Error("account repository debit internal account failed", err, logger.Fields{
			"accountNumber": accountNumber,
			"amount":        amount,
		})
		return fmt.Errorf("debit internal account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("account repository debit internal account rows affected failed", err, logger.Fields{
			"accountNumber": accountNumber,
		})
		return fmt.Errorf("debit internal account rows affected: %w", err)
	}

	if rowsAffected == 0 {
		account, getErr := r.GetByAccountNumber(ctx, accountNumber)
		if getErr != nil {
			if errors.Is(getErr, commons.ErrRecordNotFound) {
				return commons.ErrRecordNotFound
			}
			return getErr
		}
		if account.Status != domain.AccountStatusActive {
			return fmt.Errorf("account is not active")
		}
		return commons.ErrInsufficientBalance
	}

	logger.Info("account repository debit internal account success", logger.Fields{
		"accountNumber": accountNumber,
		"amount":        amount,
	})
	return nil
}

func (r *AccountRepository) CreditInternalAccount(ctx context.Context, accountNumber string, amount string) error {
	logger.Info("account repository credit internal account", logger.Fields{
		"accountNumber": accountNumber,
		"amount":        amount,
	})

	const query = `
UPDATE accounts
SET available_balance = available_balance + $2::numeric,
    updated_at = NOW()
WHERE account_number = $1
  AND status = 'ACTIVE'`

	result, err := r.db.ExecContext(ctx, query, accountNumber, amount)
	if err != nil {
		logger.Error("account repository credit internal account failed", err, logger.Fields{
			"accountNumber": accountNumber,
			"amount":        amount,
		})
		return fmt.Errorf("credit internal account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("account repository credit internal account rows affected failed", err, logger.Fields{
			"accountNumber": accountNumber,
		})
		return fmt.Errorf("credit internal account rows affected: %w", err)
	}

	if rowsAffected == 0 {
		account, getErr := r.GetByAccountNumber(ctx, accountNumber)
		if getErr != nil {
			if errors.Is(getErr, commons.ErrRecordNotFound) {
				return commons.ErrRecordNotFound
			}
			return getErr
		}
		if account.Status != domain.AccountStatusActive {
			return fmt.Errorf("account is not active")
		}
		return commons.ErrRecordNotFound
	}

	logger.Info("account repository credit internal account success", logger.Fields{
		"accountNumber": accountNumber,
		"amount":        amount,
	})
	return nil
}

func (r *AccountRepository) DepositFunds(ctx context.Context, accountNumber string, amount string) error {
	logger.Info("account repository deposit funds", logger.Fields{
		"accountNumber": accountNumber,
		"amount":        amount,
	})

	const query = `
UPDATE accounts
SET available_balance = available_balance + $2::numeric,
    ledger_balance = ledger_balance + $2::numeric,
    updated_at = NOW()
WHERE account_number = $1
  AND status = 'ACTIVE'`

	result, err := r.db.ExecContext(ctx, query, accountNumber, amount)
	if err != nil {
		logger.Error("account repository deposit funds failed", err, logger.Fields{
			"accountNumber": accountNumber,
			"amount":        amount,
		})
		return fmt.Errorf("deposit funds: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("account repository deposit funds rows affected failed", err, logger.Fields{
			"accountNumber": accountNumber,
		})
		return fmt.Errorf("deposit funds rows affected: %w", err)
	}

	if rowsAffected == 0 {
		account, getErr := r.GetByAccountNumber(ctx, accountNumber)
		if getErr != nil {
			if errors.Is(getErr, commons.ErrRecordNotFound) {
				return commons.ErrRecordNotFound
			}
			return getErr
		}
		if account.Status != domain.AccountStatusActive {
			return fmt.Errorf("account is not active")
		}
		return commons.ErrRecordNotFound
	}

	logger.Info("account repository deposit funds success", logger.Fields{
		"accountNumber": accountNumber,
		"amount":        amount,
	})
	return nil
}
