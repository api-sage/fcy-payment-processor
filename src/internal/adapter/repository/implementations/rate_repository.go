package implementations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/api-sage/ccy-payment-processor/src/internal/commons"
	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
	"github.com/api-sage/ccy-payment-processor/src/internal/logger"
)

type RateRepository struct {
	db *sql.DB
}

func NewRateRepository(db *sql.DB) *RateRepository {
	return &RateRepository{db: db}
}

func (r *RateRepository) EnsureDefaultRates(ctx context.Context) error {
	logger.Info("rate repository ensure default rates", nil)

	const query = `
INSERT INTO rates (
	from_currency,
	to_currency,
	rate,
	rate_date
) VALUES
	('USD', 'NGN', 1338.38005900, CURRENT_DATE),
	('NGN', 'USD', 0.00074717, CURRENT_DATE),
	('EUR', 'NGN', 1580.48135373, CURRENT_DATE),
	('NGN', 'EUR', 0.00063272, CURRENT_DATE),
	('GBP', 'NGN', 1810.06486117, CURRENT_DATE),
	('NGN', 'GBP', 0.00055247, CURRENT_DATE),
	('EUR', 'USD', 1.18450000, CURRENT_DATE),
	('USD', 'EUR', 0.84423808, CURRENT_DATE),
	('EUR', 'GBP', 0.87240000, CURRENT_DATE),
	('GBP', 'EUR', 1.14626318, CURRENT_DATE),
	('GBP', 'USD', 1.35774874, CURRENT_DATE),
	('USD', 'GBP', 0.73651330, CURRENT_DATE)
ON CONFLICT (from_currency, to_currency, rate_date) DO NOTHING`

	if _, err := r.db.ExecContext(ctx, query); err != nil {
		logger.Error("rate repository ensure default rates failed", err, nil)
		return fmt.Errorf("ensure default rates: %w", err)
	}

	logger.Info("rate repository ensure default rates success", nil)
	return nil
}

func (r *RateRepository) GetRates(ctx context.Context) ([]domain.Rate, error) {
	logger.Info("rate repository get rates", nil)

	const query = `
SELECT id, from_currency, to_currency, rate, rate_date, created_at
FROM rates
ORDER BY rate_date DESC, from_currency ASC, to_currency ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		logger.Error("rate repository get rates failed", err, nil)
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
			&rate.Rate,
			&rate.RateDate,
			&rate.CreatedAt,
		); err != nil {
			logger.Error("rate repository scan rate failed", err, nil)
			return nil, fmt.Errorf("scan rate: %w", err)
		}

		rates = append(rates, rate)
	}

	if err := rows.Err(); err != nil {
		logger.Error("rate repository iterate rates failed", err, nil)
		return nil, fmt.Errorf("iterate rates: %w", err)
	}

	logger.Info("rate repository get rates success", logger.Fields{
		"count": len(rates),
	})

	return rates, nil
}

func (r *RateRepository) GetRate(ctx context.Context, fromCurrency string, toCurrency string) (domain.Rate, error) {
	logger.Info("rate repository get rate", logger.Fields{
		"fromCurrency": fromCurrency,
		"toCurrency":   toCurrency,
	})

	const query = `
SELECT id, from_currency, to_currency, rate, rate_date, created_at
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
		&rate.Rate,
		&rate.RateDate,
		&rate.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info("rate repository record not found", logger.Fields{
				"fromCurrency": fromCurrency,
				"toCurrency":   toCurrency,
			})
			return domain.Rate{}, commons.ErrRecordNotFound
		}
		logger.Error("rate repository get rate failed", err, logger.Fields{
			"fromCurrency": fromCurrency,
			"toCurrency":   toCurrency,
		})
		return domain.Rate{}, fmt.Errorf("get rate: %w", err)
	}

	logger.Info("rate repository get rate success", logger.Fields{
		"rateId":       rate.ID,
		"fromCurrency": rate.FromCurrency,
		"toCurrency":   rate.ToCurrency,
		"rateDate":     rate.RateDate.Format("2006-01-02"),
	})

	return rate, nil
}
