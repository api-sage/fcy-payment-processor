package postgres

import (
	"context"
	"database/sql"
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
