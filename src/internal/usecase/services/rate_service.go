package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/repository/repo_interfaces"
	"github.com/api-sage/fcy-payment-processor/src/internal/commons"
	"github.com/api-sage/fcy-payment-processor/src/internal/domain"
	"github.com/api-sage/fcy-payment-processor/src/internal/logger"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/service_interfaces"
	"github.com/shopspring/decimal"
)

// Verify that RateService implements the service_interfaces.RateService interface
var _ service_interfaces.RateService = (*RateService)(nil)

type RateService struct {
	rateRepo repo_interfaces.RateRepository
}

func NewRateService(rateRepo repo_interfaces.RateRepository) *RateService {
	return &RateService{rateRepo: rateRepo}
}

func (s *RateService) GetRates(ctx context.Context) (commons.Response[[]models.RateResponse], error) {
	logger.Info("rate service get rates request", nil)

	rates, err := s.rateRepo.GetRates(ctx)
	if err != nil {
		logger.Error("rate service get rates failed", err, nil)
		return commons.ErrorResponse[[]models.RateResponse]("failed to get rates", "Unable to fetch rates right now"), err
	}

	resp := make([]models.RateResponse, 0, len(rates))
	for _, rate := range rates {
		resp = append(resp, mapRateToResponse(rate))
	}

	logger.Info("rate service get rates success", logger.Fields{
		"count": len(resp),
	})

	return commons.SuccessResponse("rates fetched successfully", resp), nil
}

func (s *RateService) GetRate(ctx context.Context, req models.GetRateRequest) (commons.Response[models.RateResponse], error) {
	logger.Info("rate service get rate request", logger.Fields{
		"payload": logger.SanitizePayload(req),
	})

	if err := req.Validate(); err != nil {
		logger.Error("rate service get rate validation failed", err, nil)
		return commons.ErrorResponse[models.RateResponse]("validation failed", err.Error()), err
	}

	fromCurrency := strings.ToUpper(strings.TrimSpace(req.FromCurrency))
	toCurrency := strings.ToUpper(strings.TrimSpace(req.ToCurrency))
	if fromCurrency == toCurrency {
		now := time.Now().UTC()
		response := models.RateResponse{
			ID:           0,
			FromCurrency: fromCurrency,
			ToCurrency:   toCurrency,
			Rate:         decimal.NewFromInt(1),
			RateDate:     now.Format("2006-01-02"),
			CreatedAt:    now.Format(time.RFC3339),
		}
		return commons.SuccessResponse("rate fetched successfully", response), nil
	}

	rate, err := s.rateRepo.GetRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		logger.Error("rate service get rate failed", err, logger.Fields{
			"fromCurrency": fromCurrency,
			"toCurrency":   toCurrency,
		})
		if errors.Is(err, commons.ErrRecordNotFound) {
			return commons.ErrorResponse[models.RateResponse]("Rate not found"), err
		}
		return commons.ErrorResponse[models.RateResponse]("failed to get rate", "Unable to fetch rate right now"), err
	}

	logger.Info("rate service get rate success", logger.Fields{
		"rateId":       rate.ID,
		"fromCurrency": rate.FromCurrency,
		"toCurrency":   rate.ToCurrency,
	})

	return commons.SuccessResponse("rate fetched successfully", mapRateToResponse(rate)), nil
}

func (s *RateService) ConvertRate(ctx context.Context, amount decimal.Decimal, fromCcy string, toCcy string) (decimal.Decimal, decimal.Decimal, string, error) {
	fromCurrency := strings.ToUpper(strings.TrimSpace(fromCcy))
	toCurrency := strings.ToUpper(strings.TrimSpace(toCcy))

	if fromCurrency == "" {
		return decimal.Decimal{}, decimal.Decimal{}, "", fmt.Errorf("fromCcy is required")
	}
	if toCurrency == "" {
		return decimal.Decimal{}, decimal.Decimal{}, "", fmt.Errorf("toCcy is required")
	}
	if len(fromCurrency) != 3 || len(toCurrency) != 3 {
		return decimal.Decimal{}, decimal.Decimal{}, "", fmt.Errorf("fromCcy and toCcy must be 3 characters")
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Decimal{}, decimal.Decimal{}, "", fmt.Errorf("amount must be greater than zero")
	}
	if fromCurrency == toCurrency {
		return amount, decimal.NewFromInt(1), time.Now().UTC().Format("2006-01-02"), nil
	}

	rate, err := s.rateRepo.GetRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		return decimal.Decimal{}, decimal.Decimal{}, "", err
	}
	usedRate := rate.Rate
	if usedRate.Equal(decimal.Zero) {
		return decimal.Decimal{}, decimal.Decimal{}, "", fmt.Errorf("rate cannot be zero")
	}

	converted := amount.Mul(usedRate)
	return converted, usedRate, rate.RateDate.Format("2006-01-02"), nil
}

func (s *RateService) GetCcyRates(ctx context.Context, req models.GetCcyRatesRequest) (commons.Response[models.GetCcyRatesResponse], error) {
	logger.Info("rate service convert fcy amount request", logger.Fields{
		"payload": logger.SanitizePayload(req),
	})

	if err := req.Validate(); err != nil {
		logger.Error("rate service convert fcy amount validation failed", err, nil)
		return commons.ErrorResponse[models.GetCcyRatesResponse]("validation failed", err.Error()), err
	}

	convertedAmount, rateUsed, rateDate, err := s.ConvertRate(ctx, req.Amount, req.FromCcy, req.ToCcy)
	if err != nil {
		logger.Error("rate service convert fcy amount failed", err, logger.Fields{
			"fromCcy": req.FromCcy,
			"toCcy":   req.ToCcy,
		})
		if errors.Is(err, commons.ErrRecordNotFound) {
			return commons.ErrorResponse[models.GetCcyRatesResponse]("Rate not found for currency pair", "Rate not found for currency pair"), err
		}
		return commons.ErrorResponse[models.GetCcyRatesResponse]("failed to get currency rates", "Unable to fetch currency rates right now"), err
	}

	response := models.GetCcyRatesResponse{
		Amount:          req.Amount,
		FromCcy:         strings.ToUpper(strings.TrimSpace(req.FromCcy)),
		ToCcy:           strings.ToUpper(strings.TrimSpace(req.ToCcy)),
		ConvertedAmount: convertedAmount,
		RateUsed:        rateUsed,
		RateDate:        rateDate,
	}

	logger.Info("rate service convert fcy amount success", logger.Fields{
		"fromCcy":         response.FromCcy,
		"toCcy":           response.ToCcy,
		"convertedAmount": response.ConvertedAmount,
		"rateDate":        response.RateDate,
	})

	return commons.SuccessResponse("currency rate fetched successfully", response), nil
}

func mapRateToResponse(rate domain.Rate) models.RateResponse {
	return models.RateResponse{
		ID:           rate.ID,
		FromCurrency: rate.FromCurrency,
		ToCurrency:   rate.ToCurrency,
		Rate:         rate.Rate,
		RateDate:     rate.RateDate.Format("2006-01-02"),
		CreatedAt:    rate.CreatedAt.Format(time.RFC3339),
	}
}
