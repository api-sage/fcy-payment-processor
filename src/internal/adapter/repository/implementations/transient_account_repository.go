package implementations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/api-sage/ccy-payment-processor/src/internal/commons"
	"github.com/api-sage/ccy-payment-processor/src/internal/logger"
)

type TransientAccountRepository struct {
	db *sql.DB
}

func NewTransientAccountRepository(db *sql.DB) *TransientAccountRepository {
	return &TransientAccountRepository{db: db}
}

func (r *TransientAccountRepository) EnsureInternalAccounts(
	ctx context.Context,
	internalTransientAccountNumber string,
	internalChargesAccountNumber string,
	internalVATAccountNumber string,
	externalUSDGLAccountNumber string,
	externalGBPGLAccountNumber string,
	externalEURGLAccountNumber string,
	externalNGNGLAccountNumber string,

) error {
	logger.Info("transient account repository ensure internal accounts", logger.Fields{
		"internalTransientAccountNumber": internalTransientAccountNumber,
		"internalChargesAccountNumber":   internalChargesAccountNumber,
		"internalVATAccountNumber":       internalVATAccountNumber,
		"externalUSDGLAccountNumber":     externalUSDGLAccountNumber,
		"externalGBPGLAccountNumber":     externalGBPGLAccountNumber,
		"externalEURGLAccountNumber":     externalEURGLAccountNumber,
		"externalNGNGLAccountNumber":     externalNGNGLAccountNumber,
	})

	const query = `
INSERT INTO transient_accounts (
	account_number,
	account_description,
	currency,
	available_balance
) VALUES
	($1, 'Internal Transient Account', 'MCY', 0.00),
	($2, 'Internal Charges Account', 'USD', 0.00),
	($3, 'Internal VAT Account', 'USD', 0.00),
	($4, 'External USD GL Account', 'USD', 0.00),
	($5, 'External GBP GL Account', 'GBP', 0.00),
	($6, 'External EUR GL Account', 'EUR', 0.00),
	($7, 'External NGN GL Account', 'NGN', 0.00)
ON CONFLICT (account_number) DO NOTHING`

	if _, err := r.db.ExecContext(
		ctx,
		query,
		internalTransientAccountNumber,
		internalChargesAccountNumber,
		internalVATAccountNumber,
		externalUSDGLAccountNumber,
		externalGBPGLAccountNumber,
		externalEURGLAccountNumber,
		externalNGNGLAccountNumber,
	); err != nil {
		logger.Error("transient account repository ensure internal accounts failed", err, nil)
		return fmt.Errorf("ensure internal transient accounts: %w", err)
	}

	logger.Info("transient account repository ensure internal accounts success", nil)
	return nil
}

func (r *TransientAccountRepository) DebitSuspenseAccount(ctx context.Context, suspenseAccountNumber string, currency string, amount string) error {
	logger.Info("transient account repository debit", logger.Fields{
		"accountNumber": suspenseAccountNumber,
		"currency":      currency,
		"amount":        amount,
	})

	const existsQuery = `
SELECT 1
FROM transient_accounts
WHERE account_number = $1
  AND UPPER(currency) = UPPER($2)`

	var exists int
	if err := r.db.QueryRowContext(ctx, existsQuery, suspenseAccountNumber, currency).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info("transient account repository record not found", logger.Fields{
				"accountNumber": suspenseAccountNumber,
				"currency":      currency,
			})
			return commons.ErrRecordNotFound
		}
		logger.Error("transient account repository check failed", err, logger.Fields{
			"accountNumber": suspenseAccountNumber,
			"currency":      currency,
		})
		return fmt.Errorf("check transient account: %w", err)
	}

	const debitQuery = `
UPDATE transient_accounts
SET available_balance = available_balance - $2::numeric,
    updated_at = NOW()
WHERE account_number = $1
  AND available_balance >= $2::numeric`

	result, err := r.db.ExecContext(ctx, debitQuery, suspenseAccountNumber, amount)
	if err != nil {
		logger.Error("transient account repository debit failed", err, logger.Fields{
			"accountNumber": suspenseAccountNumber,
			"currency":      currency,
			"amount":        amount,
		})
		return fmt.Errorf("debit transient account: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		logger.Error("transient account repository debit rows affected failed", err, logger.Fields{
			"accountNumber": suspenseAccountNumber,
			"currency":      currency,
		})
		return fmt.Errorf("debit transient account rows affected: %w", err)
	}
	if rows == 0 {
		logger.Info("transient account repository insufficient balance", logger.Fields{
			"accountNumber": suspenseAccountNumber,
			"currency":      currency,
			"amount":        amount,
		})
		return commons.ErrInsufficientBalance
	}

	logger.Info("transient account repository debit success", logger.Fields{
		"accountNumber": suspenseAccountNumber,
		"currency":      currency,
		"amount":        amount,
	})
	return nil
}

func (r *TransientAccountRepository) CreditSuspenseAccount(ctx context.Context, suspenseAccountNumber string, currency string, amount string) error {
	logger.Info("transient account repository credit", logger.Fields{
		"accountNumber": suspenseAccountNumber,
		"currency":      currency,
		"amount":        amount,
	})

	const query = `
UPDATE transient_accounts
SET available_balance = available_balance + $2::numeric,
    updated_at = NOW()
WHERE account_number = $1
  AND UPPER(currency) = UPPER($3)`

	result, err := r.db.ExecContext(ctx, query, suspenseAccountNumber, amount, currency)
	if err != nil {
		logger.Error("transient account repository credit failed", err, logger.Fields{
			"accountNumber": suspenseAccountNumber,
			"currency":      currency,
			"amount":        amount,
		})
		return fmt.Errorf("credit transient account: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		logger.Error("transient account repository credit rows affected failed", err, logger.Fields{
			"accountNumber": suspenseAccountNumber,
			"currency":      currency,
		})
		return fmt.Errorf("credit transient account rows affected: %w", err)
	}
	if rows == 0 {
		logger.Info("transient account repository record not found", logger.Fields{
			"accountNumber": suspenseAccountNumber,
			"currency":      currency,
		})
		return commons.ErrRecordNotFound
	}

	logger.Info("transient account repository credit success", logger.Fields{
		"accountNumber": suspenseAccountNumber,
		"currency":      currency,
		"amount":        amount,
	})
	return nil
}

func (r *TransientAccountRepository) SettleFromSuspenseToFees(
	ctx context.Context,
	suspenseAccountNumber string,
	chargeAmount string,
	vatAmount string,
	chargesAccountNumber string,
	vatAccountNumber string,
	chargeUSD string,
	vatUSD string,
) error {
	logger.Info("transient account repository settle from suspense to fees", logger.Fields{
		"suspenseAccountNumber": suspenseAccountNumber,
		"chargeAmount":          chargeAmount,
		"vatAmount":             vatAmount,
		"chargesAccountNumber":  chargesAccountNumber,
		"vatAccountNumber":      vatAccountNumber,
		"chargeUSD":             chargeUSD,
		"vatUSD":                vatUSD,
	})

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin settlement transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Debit suspense by combined charge and VAT before crediting fee accounts.
	debitSuspenseSumQuery := `
UPDATE transient_accounts
SET available_balance = available_balance - ($2::numeric + $3::numeric),
    updated_at = NOW()
WHERE account_number = $1
  AND available_balance >= ($2::numeric + $3::numeric)`
	result, execErr := tx.ExecContext(ctx, debitSuspenseSumQuery, suspenseAccountNumber, chargeAmount, vatAmount)
	if execErr != nil {
		err = fmt.Errorf("debit suspense for settlement: %w", execErr)
		return err
	}
	rows, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		err = fmt.Errorf("debit suspense rows affected: %w", rowsErr)
		return err
	}
	if rows == 0 {
		err = commons.ErrInsufficientBalance
		return err
	}

	creditFeeQuery := `
UPDATE transient_accounts
SET available_balance = available_balance + $2::numeric,
    updated_at = NOW()
WHERE account_number = $1
  AND UPPER(currency) = 'USD'`

	result, execErr = tx.ExecContext(ctx, creditFeeQuery, chargesAccountNumber, chargeUSD)
	if execErr != nil {
		err = fmt.Errorf("credit charges account for settlement: %w", execErr)
		return err
	}
	rows, rowsErr = result.RowsAffected()
	if rowsErr != nil {
		err = fmt.Errorf("credit charges rows affected: %w", rowsErr)
		return err
	}
	if rows == 0 {
		err = commons.ErrRecordNotFound
		return err
	}

	result, execErr = tx.ExecContext(ctx, creditFeeQuery, vatAccountNumber, vatUSD)
	if execErr != nil {
		err = fmt.Errorf("credit vat account for settlement: %w", execErr)
		return err
	}
	rows, rowsErr = result.RowsAffected()
	if rowsErr != nil {
		err = fmt.Errorf("credit vat rows affected: %w", rowsErr)
		return err
	}
	if rows == 0 {
		err = commons.ErrRecordNotFound
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit settlement transaction: %w", err)
	}

	logger.Info("transient account repository settle from suspense to fees success", logger.Fields{
		"suspenseAccountNumber": suspenseAccountNumber,
		"chargesAccountNumber":  chargesAccountNumber,
		"vatAccountNumber":      vatAccountNumber,
	})

	return nil
}
