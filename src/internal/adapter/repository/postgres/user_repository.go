package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
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
	transaction_pin_has
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, customer_id, first_name, middle_name, last_name, dob, phone_number, id_type, id_number, kyc_level, transaction_pin_has, created_at, updated_at`

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
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	return created, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	const query = `
SELECT id, customer_id, first_name, middle_name, last_name, dob, phone_number, id_type, id_number, kyc_level, transaction_pin_has, created_at, updated_at
FROM users
WHERE id = $1`

	var user domain.User
	if err := scanUser(r.db.QueryRowContext(ctx, query, id), &user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, fmt.Errorf("user not found: %w", err)
		}
		return domain.User{}, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user domain.User) (domain.User, error) {
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
	transaction_pin_has = $11,
	updated_at = NOW()
WHERE id = $1
RETURNING id, customer_id, first_name, middle_name, last_name, dob, phone_number, id_type, id_number, kyc_level, transaction_pin_has, created_at, updated_at`

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
			return domain.User{}, fmt.Errorf("user not found: %w", err)
		}
		return domain.User{}, fmt.Errorf("update user: %w", err)
	}

	return updated, nil
}

func (r *UserRepository) GetTransactionPinHashByCustomerID(ctx context.Context, customerID string) (string, error) {
	const query = `
SELECT transaction_pin_has
FROM users
WHERE customer_id = $1`

	var transactionPinHash string
	if err := r.db.QueryRowContext(ctx, query, customerID).Scan(&transactionPinHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("user not found: %w", err)
		}
		return "", fmt.Errorf("get transaction pin hash by customer id: %w", err)
	}

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
