package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
)

type RateRepository struct {
	db *sql.DB
}

func NewRateRepository(db *sql.DB) *RateRepository {
	return &RateRepository{db: db}
}

func (r *RateRepository) GetRates(ctx context.Context) ([]domain.Rate, error) {
	const query = `
SELECT id, from_currency, to_currency, sell_rate, buy_rate, rate_date, created_at
FROM rates
ORDER BY rate_date DESC, from_currency ASC, to_currency ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get rates: %w", err)
	}
	defer rows.Close()

	rates := make([]domain.Rate, 0)
	for rows.Next() {
		var rate domain.Rate
		if err := rows.Scan(
			&rate.ID,
			&rate.FromCurrency,
			&rate.ToCurrency,
			&rate.SellRate,
			&rate.BuyRate,
			&rate.RateDate,
			&rate.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan rate: %w", err)
		}

		rates = append(rates, rate)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rates: %w", err)
	}

	return rates, nil
}

func (r *RateRepository) GetRate(ctx context.Context, fromCurrency string, toCurrency string) (domain.Rate, error) {
	const query = `
SELECT id, from_currency, to_currency, sell_rate, buy_rate, rate_date, created_at
FROM rates
WHERE from_currency = $1
  AND to_currency = $2
ORDER BY rate_date DESC
LIMIT 1`

	var rate domain.Rate
	if err := r.db.QueryRowContext(ctx, query, fromCurrency, toCurrency).Scan(
		&rate.ID,
		&rate.FromCurrency,
		&rate.ToCurrency,
		&rate.SellRate,
		&rate.BuyRate,
		&rate.RateDate,
		&rate.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Rate{}, domain.ErrRecordNotFound
		}
		return domain.Rate{}, fmt.Errorf("get rate: %w", err)
	}

	return rate, nil
}
