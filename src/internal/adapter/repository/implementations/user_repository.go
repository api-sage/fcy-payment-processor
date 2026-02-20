package implementations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
	"github.com/api-sage/ccy-payment-processor/src/internal/logger"
)

type UserRepository struct {
	db *sql.DB
}

type rowScanner interface {
	Scan(dest ...any) error
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	logger.Info("user repository create", logger.Fields{
		"customerId": user.CustomerID,
		"firstName":  user.FirstName,
		"lastName":   user.LastName,
	})

	const query = `
INSERT INTO users (
	customer_id,
	first_name,
	middle_name,
	last_name,
	dob,
	phone_number,
	id_type,
	id_number,
	kyc_level,
	transaction_pin_hash
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, customer_id, first_name, middle_name, last_name, dob, phone_number, id_type, id_number, kyc_level, transaction_pin_hash, created_at, updated_at`

	var created domain.User
	if err := scanUser(r.db.QueryRowContext(
		ctx,
		query,
		user.CustomerID,
		user.FirstName,
		user.MiddleName,
		user.LastName,
		user.DOB,
		user.PhoneNumber,
		user.IDType,
		user.IDNumber,
		user.KYCLevel,
		user.TransactionPinHash,
	), &created); err != nil {
		logger.Error("user repository create failed", err, logger.Fields{
			"customerId": user.CustomerID,
		})
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	logger.Info("user repository create success", logger.Fields{
		"userId":      created.ID,
		"customerId":  created.CustomerID,
		"transaction": "create",
	})

	return created, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	logger.Info("user repository get by id", logger.Fields{
		"userId": id,
	})

	const query = `
SELECT id, customer_id, first_name, middle_name, last_name, dob, phone_number, id_type, id_number, kyc_level, transaction_pin_hash, created_at, updated_at
FROM users
WHERE id = $1`

	var user domain.User
	if err := scanUser(r.db.QueryRowContext(ctx, query, id), &user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info("user repository record not found", logger.Fields{
				"userId": id,
			})
			return domain.User{}, domain.ErrRecordNotFound
		}
		logger.Error("user repository get by id failed", err, logger.Fields{
			"userId": id,
		})
		return domain.User{}, fmt.Errorf("get user by id: %w", err)
	}

	logger.Info("user repository get by id success", logger.Fields{
		"userId":     user.ID,
		"customerId": user.CustomerID,
	})

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user domain.User) (domain.User, error) {
	logger.Info("user repository update", logger.Fields{
		"userId":     user.ID,
		"customerId": user.CustomerID,
	})

	const query = `
UPDATE users
SET customer_id = $2,
	first_name = $3,
	middle_name = $4,
	last_name = $5,
	dob = $6,
	phone_number = $7,
	id_type = $8,
	id_number = $9,
	kyc_level = $10,
	transaction_pin_hash = $11,
	updated_at = NOW()
WHERE id = $1
RETURNING id, customer_id, first_name, middle_name, last_name, dob, phone_number, id_type, id_number, kyc_level, transaction_pin_hash, created_at, updated_at`

	var updated domain.User
	if err := scanUser(r.db.QueryRowContext(
		ctx,
		query,
		user.ID,
		user.CustomerID,
		user.FirstName,
		user.MiddleName,
		user.LastName,
		user.DOB,
		user.PhoneNumber,
		user.IDType,
		user.IDNumber,
		user.KYCLevel,
		user.TransactionPinHash,
	), &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info("user repository record not found for update", logger.Fields{
				"userId": user.ID,
			})
			return domain.User{}, domain.ErrRecordNotFound
		}
		logger.Error("user repository update failed", err, logger.Fields{
			"userId": user.ID,
		})
		return domain.User{}, fmt.Errorf("update user: %w", err)
	}

	logger.Info("user repository update success", logger.Fields{
		"userId":     updated.ID,
		"customerId": updated.CustomerID,
	})

	return updated, nil
}

func (r *UserRepository) GetTransactionPinHashByCustomerID(ctx context.Context, customerID string) (string, error) {
	logger.Info("user repository get pin hash by customer id", logger.Fields{
		"customerId": customerID,
	})

	const query = `
SELECT transaction_pin_hash
FROM users
WHERE customer_id = $1`

	var transactionPinHash string
	if err := r.db.QueryRowContext(ctx, query, customerID).Scan(&transactionPinHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info("user repository pin hash record not found", logger.Fields{
				"customerId": customerID,
			})
			return "", domain.ErrRecordNotFound
		}
		logger.Error("user repository get pin hash failed", err, logger.Fields{
			"customerId": customerID,
		})
		return "", fmt.Errorf("get transaction pin hash by customer id: %w", err)
	}

	logger.Info("user repository get pin hash success", logger.Fields{
		"customerId": customerID,
	})

	return transactionPinHash, nil
}

func scanUser(row rowScanner, user *domain.User) error {
	return row.Scan(
		&user.ID,
		&user.CustomerID,
		&user.FirstName,
		&user.MiddleName,
		&user.LastName,
		&user.DOB,
		&user.PhoneNumber,
		&user.IDType,
		&user.IDNumber,
		&user.KYCLevel,
		&user.TransactionPinHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
}
