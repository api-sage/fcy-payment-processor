package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
	"github.com/api-sage/ccy-payment-processor/src/internal/logger"
)

type RateRepository struct {
	db *sql.DB
}

func NewRateRepository(db *sql.DB) *RateRepository {
	return &RateRepository{db: db}
}

func (r *RateRepository) GetRates(ctx context.Context) ([]domain.Rate, error) {
	logger.Info("rate repository get rates", nil)

	const query = `
SELECT id, from_currency, to_currency, sell_rate, buy_rate, rate_date, created_at
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
			&rate.SellRate,
			&rate.BuyRate,
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
			logger.Info("rate repository record not found", logger.Fields{
				"fromCurrency": fromCurrency,
				"toCurrency":   toCurrency,
			})
			return domain.Rate{}, domain.ErrRecordNotFound
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
