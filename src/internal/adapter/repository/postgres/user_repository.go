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
		user.TransactionPinHas,
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
		&user.TransactionPinHas,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
}
