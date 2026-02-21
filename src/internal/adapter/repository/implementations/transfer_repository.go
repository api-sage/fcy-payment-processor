package implementations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/api-sage/ccy-payment-processor/src/internal/commons"
	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
	"github.com/api-sage/ccy-payment-processor/src/internal/logger"
)

type TransferRepository struct {
	db *sql.DB
}

func NewTransferRepository(db *sql.DB) *TransferRepository {
	return &TransferRepository{db: db}
}

func (r *TransferRepository) Create(ctx context.Context, transfer domain.Transfer) (domain.Transfer, error) {
	logger.Info("transfer repository create", logger.Fields{
		"transactionReference": transfer.TransactionReference,
		"debitAccountNumber":   transfer.DebitAccountNumber,
		"creditAccountNumber":  transfer.CreditAccountNumber,
		"status":               transfer.Status,
	})

	const query = `
INSERT INTO transfers (
	external_refernece,
	transaction_reference,
	debit_account_number,
	credit_account_number,
	beneficiary_bank_code,
	debit_currency,
	credit_currency,
	debit_amount,
	credit_amount,
	fcy_rate,
	charge_amount,
	vat_amount,
	narration,
	status,
	audit_payload
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
)
RETURNING id, created_at, updated_at, processed_at`

	var (
		id          string
		createdAt   time.Time
		updatedAt   time.Time
		processedAt sql.NullTime
	)

	if err := r.db.QueryRowContext(
		ctx,
		query,
		transfer.ExternalRefernece,
		transfer.TransactionReference,
		transfer.DebitAccountNumber,
		transfer.CreditAccountNumber,
		transfer.BeneficiaryBankCode,
		transfer.DebitCurrency,
		transfer.CreditCurrency,
		transfer.DebitAmount,
		transfer.CreditAmount,
		transfer.FCYRate,
		transfer.ChargeAmount,
		transfer.VATAmount,
		transfer.Narration,
		transfer.Status,
		transfer.AuditPayload,
	).Scan(&id, &createdAt, &updatedAt, &processedAt); err != nil {
		logger.Error("transfer repository create failed", err, logger.Fields{
			"transactionReference": transfer.TransactionReference,
		})
		return domain.Transfer{}, fmt.Errorf("create transfer: %w", err)
	}

	transfer.ID = id
	transfer.CreatedAt = createdAt
	transfer.UpdatedAt = updatedAt
	if processedAt.Valid {
		value := processedAt.Time
		transfer.ProcessedAt = &value
	}

	logger.Info("transfer repository create success", logger.Fields{
		"transferId":           transfer.ID,
		"transactionReference": transfer.TransactionReference,
	})

	return transfer, nil
}

func (r *TransferRepository) Update(ctx context.Context, transfer domain.Transfer) (domain.Transfer, error) {
	logger.Info("transfer repository update", logger.Fields{
		"transferId":           transfer.ID,
		"transactionReference": transfer.TransactionReference,
		"status":               transfer.Status,
	})

	const query = `
UPDATE transfers
SET external_refernece = $2,
    transaction_reference = $3,
    debit_account_number = $4,
    credit_account_number = $5,
    beneficiary_bank_code = $6,
    debit_currency = $7,
    credit_currency = $8,
    debit_amount = $9,
    credit_amount = $10,
    fcy_rate = $11,
    charge_amount = $12,
    vat_amount = $13,
    narration = $14,
    status = $15,
    audit_payload = $16,
    updated_at = NOW(),
    processed_at = CASE
        WHEN $15 IN ('SUCCESS', 'FAILED', 'CLOSED') THEN NOW()
        ELSE processed_at
    END
WHERE id = $1
RETURNING created_at, updated_at, processed_at`

	var (
		createdAt   time.Time
		updatedAt   time.Time
		processedAt sql.NullTime
	)

	if err := r.db.QueryRowContext(
		ctx,
		query,
		transfer.ID,
		transfer.ExternalRefernece,
		transfer.TransactionReference,
		transfer.DebitAccountNumber,
		transfer.CreditAccountNumber,
		transfer.BeneficiaryBankCode,
		transfer.DebitCurrency,
		transfer.CreditCurrency,
		transfer.DebitAmount,
		transfer.CreditAmount,
		transfer.FCYRate,
		transfer.ChargeAmount,
		transfer.VATAmount,
		transfer.Narration,
		transfer.Status,
		transfer.AuditPayload,
	).Scan(&createdAt, &updatedAt, &processedAt); err != nil {
		if err == sql.ErrNoRows {
			logger.Info("transfer repository record not found", logger.Fields{
				"transferId": transfer.ID,
			})
			return domain.Transfer{}, commons.ErrRecordNotFound
		}
		logger.Error("transfer repository update failed", err, logger.Fields{
			"transferId": transfer.ID,
		})
		return domain.Transfer{}, fmt.Errorf("update transfer: %w", err)
	}

	transfer.CreatedAt = createdAt
	transfer.UpdatedAt = updatedAt
	transfer.ProcessedAt = nil
	if processedAt.Valid {
		value := processedAt.Time
		transfer.ProcessedAt = &value
	}

	logger.Info("transfer repository update success", logger.Fields{
		"transferId": transfer.ID,
		"status":     transfer.Status,
	})

	return transfer, nil
}

func (r *TransferRepository) Get(ctx context.Context, id string, transactionReference string, externalRefernece string) (domain.Transfer, error) {
	trimmedID := strings.TrimSpace(id)
	trimmedTxRef := strings.TrimSpace(transactionReference)
	trimmedExternalRef := strings.TrimSpace(externalRefernece)

	if trimmedID == "" && trimmedTxRef == "" && trimmedExternalRef == "" {
		return domain.Transfer{}, fmt.Errorf("id or transactionReference or externalRefernece is required")
	}

	logger.Info("transfer repository get", logger.Fields{
		"id":                   trimmedID,
		"transactionReference": trimmedTxRef,
		"externalRefernece":    trimmedExternalRef,
	})

	const query = `
SELECT id,
       external_refernece,
       transaction_reference,
       debit_account_number,
       credit_account_number,
       beneficiary_bank_code,
       debit_currency,
       credit_currency,
       debit_amount,
       credit_amount,
       fcy_rate,
       charge_amount,
       vat_amount,
       narration,
       status,
       audit_payload,
       created_at,
       updated_at,
       processed_at
FROM transfers
WHERE ($1 <> '' AND id::text = $1)
   OR ($2 <> '' AND transaction_reference = $2)
   OR ($3 <> '' AND external_refernece = $3)
ORDER BY updated_at DESC
LIMIT 1`

	var (
		transfer               domain.Transfer
		externalReference      sql.NullString
		transactionReferenceDB sql.NullString
		creditAccountNumber    sql.NullString
		beneficiaryBankCode    sql.NullString
		narration              sql.NullString
		processedAt            sql.NullTime
	)

	if err := r.db.QueryRowContext(ctx, query, trimmedID, trimmedTxRef, trimmedExternalRef).Scan(
		&transfer.ID,
		&externalReference,
		&transactionReferenceDB,
		&transfer.DebitAccountNumber,
		&creditAccountNumber,
		&beneficiaryBankCode,
		&transfer.DebitCurrency,
		&transfer.CreditCurrency,
		&transfer.DebitAmount,
		&transfer.CreditAmount,
		&transfer.FCYRate,
		&transfer.ChargeAmount,
		&transfer.VATAmount,
		&narration,
		&transfer.Status,
		&transfer.AuditPayload,
		&transfer.CreatedAt,
		&transfer.UpdatedAt,
		&processedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			logger.Info("transfer repository record not found", logger.Fields{
				"id":                   trimmedID,
				"transactionReference": trimmedTxRef,
				"externalRefernece":    trimmedExternalRef,
			})
			return domain.Transfer{}, commons.ErrRecordNotFound
		}
		logger.Error("transfer repository get failed", err, logger.Fields{
			"id":                   trimmedID,
			"transactionReference": trimmedTxRef,
			"externalRefernece":    trimmedExternalRef,
		})
		return domain.Transfer{}, fmt.Errorf("get transfer: %w", err)
	}

	if externalReference.Valid {
		value := externalReference.String
		transfer.ExternalRefernece = &value
	}
	if transactionReferenceDB.Valid {
		value := transactionReferenceDB.String
		transfer.TransactionReference = &value
	}
	if creditAccountNumber.Valid {
		value := creditAccountNumber.String
		transfer.CreditAccountNumber = &value
	}
	if beneficiaryBankCode.Valid {
		value := beneficiaryBankCode.String
		transfer.BeneficiaryBankCode = &value
	}
	if narration.Valid {
		value := narration.String
		transfer.Narration = &value
	}
	if processedAt.Valid {
		value := processedAt.Time
		transfer.ProcessedAt = &value
	}

	logger.Info("transfer repository get success", logger.Fields{
		"transferId":           transfer.ID,
		"transactionReference": transfer.TransactionReference,
		"status":               transfer.Status,
	})

	return transfer, nil
}

func (r *TransferRepository) ProcessInternalTransfer(ctx context.Context, debitAccountNumber string, debitAmount string, suspenseAccountNumber string, debitSuspenseAccountAmount string, creditAccountNumber string, creditAmount string) error {
	logger.Info("transfer repository process internal transfer", logger.Fields{
		"debitAccountNumber":         debitAccountNumber,
		"debitAmount":                debitAmount,
		"suspenseAccountNumber":      suspenseAccountNumber,
		"debitSuspenseAccountAmount": debitSuspenseAccountAmount,
		"creditAccountNumber":        creditAccountNumber,
		"creditAmount":               creditAmount,
	})

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("transfer repository begin tx failed", err, nil)
		return fmt.Errorf("begin transfer transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	debitSenderQuery := `
UPDATE accounts
SET available_balance = available_balance - $2::numeric,
    updated_at = NOW()
WHERE account_number = $1
  AND status = 'ACTIVE'
  AND available_balance >= $2::numeric`
	if _, err = execRequiredRows(ctx, tx, debitSenderQuery, debitAccountNumber, debitAmount); err != nil {
		return err
	}

	creditSuspenseQuery := `
UPDATE transient_accounts
SET available_balance = available_balance + $2::numeric,
    updated_at = NOW()
WHERE account_number = $1`
	if _, err = execRequiredRows(ctx, tx, creditSuspenseQuery, suspenseAccountNumber, debitAmount); err != nil {
		return err
	}

	debitSuspenseQuery := `
UPDATE transient_accounts
SET available_balance = available_balance - $2::numeric,
    updated_at = NOW()
WHERE account_number = $1
  AND available_balance >= $2::numeric`
	if _, err = execRequiredRows(ctx, tx, debitSuspenseQuery, suspenseAccountNumber, debitSuspenseAccountAmount); err != nil {
		return err
	}

	creditBeneficiaryQuery := `
UPDATE accounts
SET available_balance = available_balance + $2::numeric,
    updated_at = NOW()
WHERE account_number = $1
  AND status = 'ACTIVE'`
	if _, err = execRequiredRows(ctx, tx, creditBeneficiaryQuery, creditAccountNumber, creditAmount); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		logger.Error("transfer repository commit tx failed", err, nil)
		return fmt.Errorf("commit transfer transaction: %w", err)
	}

	logger.Info("transfer repository process internal transfer success", logger.Fields{
		"debitAccountNumber":  debitAccountNumber,
		"creditAccountNumber": creditAccountNumber,
	})
	return nil
}

func (r *TransferRepository) UpdateStatus(ctx context.Context, transferID string, status domain.TransferStatus) error {
	logger.Info("transfer repository update status", logger.Fields{
		"transferId": transferID,
		"status":     status,
	})

	const query = `
UPDATE transfers
SET status = $2::varchar,
    updated_at = NOW(),
    processed_at = CASE
        WHEN $2::varchar IN ('SUCCESS', 'FAILED', 'CLOSED') THEN NOW()
        ELSE processed_at
    END
WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, transferID, status)
	if err != nil {
		logger.Error("transfer repository update status failed", err, logger.Fields{
			"transferId": transferID,
			"status":     status,
		})
		return fmt.Errorf("update transfer status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update transfer status rows affected: %w", err)
	}
	if rows == 0 {
		return commons.ErrRecordNotFound
	}

	logger.Info("transfer repository update status success", logger.Fields{
		"transferId": transferID,
		"status":     status,
	})
	return nil
}

func execRequiredRows(ctx context.Context, tx *sql.Tx, query string, args ...any) (int64, error) {
	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("execute transaction statement: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read rows affected: %w", err)
	}
	if rows == 0 {
		return 0, errors.New("transaction posting failed: record not found, inactive, or insufficient balance")
	}
	return rows, nil
}
