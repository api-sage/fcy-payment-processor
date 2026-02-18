package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
)

type UserRepository struct {
	db *sql.DB
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
	if err := r.db.QueryRowContext(
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
	).Scan(
		&created.ID,
		&created.CustomerID,
		&created.FirstName,
		&created.MiddleName,
		&created.LastName,
		&created.DOB,
		&created.PhoneNumber,
		&created.IDType,
		&created.IDNumber,
		&created.KYCLevel,
		&created.TransactionPinHas,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	return created, nil
}
