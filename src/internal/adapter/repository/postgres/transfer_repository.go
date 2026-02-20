package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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
			return domain.Transfer{}, domain.ErrRecordNotFound
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
	return domain.Transfer{}, fmt.Errorf("not implemented")
}

func (r *TransferRepository) ProcessInternalTransfer(ctx context.Context, debitAccountNumber string, debitAmount string, suspenseAccountNumber string, creditAccountNumber string, creditAmount string) error {
	return fmt.Errorf("not implemented")
}

func (r *TransferRepository) UpdateStatus(ctx context.Context, transferID string, status domain.TransferStatus) error {
	return fmt.Errorf("not implemented")
}
