package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
	"github.com/api-sage/ccy-payment-processor/src/internal/logger"
	"github.com/shopspring/decimal"
)

type RateService struct {
	rateRepo domain.RateRepository
}

func NewRateService(rateRepo domain.RateRepository) *RateService {
	return &RateService{rateRepo: rateRepo}
}

func (s *RateService) GetRates(ctx context.Context) (models.Response[[]models.RateResponse], error) {
	logger.Info("rate service get rates request", nil)

	rates, err := s.rateRepo.GetRates(ctx)
	if err != nil {
		logger.Error("rate service get rates failed", err, nil)
		return models.ErrorResponse[[]models.RateResponse]("failed to get rates", "Unable to fetch rates right now"), err
	}

	resp := make([]models.RateResponse, 0, len(rates))
	for _, rate := range rates {
		resp = append(resp, mapRateToResponse(rate))
	}

	logger.Info("rate service get rates success", logger.Fields{
		"count": len(resp),
	})

	return models.SuccessResponse("rates fetched successfully", resp), nil
}

func (s *RateService) GetRate(ctx context.Context, req models.GetRateRequest) (models.Response[models.RateResponse], error) {
	logger.Info("rate service get rate request", logger.Fields{
		"payload": logger.SanitizePayload(req),
	})

	if err := req.Validate(); err != nil {
		logger.Error("rate service get rate validation failed", err, nil)
		return models.ErrorResponse[models.RateResponse]("validation failed", err.Error()), err
	}

	fromCurrency := strings.ToUpper(strings.TrimSpace(req.FromCurrency))
	toCurrency := strings.ToUpper(strings.TrimSpace(req.ToCurrency))

	rate, err := s.rateRepo.GetRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		logger.Error("rate service get rate failed", err, logger.Fields{
			"fromCurrency": fromCurrency,
			"toCurrency":   toCurrency,
		})
		if errors.Is(err, domain.ErrRecordNotFound) {
			return models.ErrorResponse[models.RateResponse]("Rate not found"), err
		}
		return models.ErrorResponse[models.RateResponse]("failed to get rate", "Unable to fetch rate right now"), err
	}

	logger.Info("rate service get rate success", logger.Fields{
		"rateId":       rate.ID,
		"fromCurrency": rate.FromCurrency,
		"toCurrency":   rate.ToCurrency,
	})

	return models.SuccessResponse("rate fetched successfully", mapRateToResponse(rate)), nil
}

func (s *RateService) ConvertRate(ctx context.Context, amount string, fromCcy string, toCcy string) (string, string, string, error) {
	trimmedAmount := strings.TrimSpace(amount)
	fromCurrency := strings.ToUpper(strings.TrimSpace(fromCcy))
	toCurrency := strings.ToUpper(strings.TrimSpace(toCcy))

	if trimmedAmount == "" {
		return "", "", "", fmt.Errorf("amount is required")
	}
	if fromCurrency == "" {
		return "", "", "", fmt.Errorf("fromCcy is required")
	}
	if toCurrency == "" {
		return "", "", "", fmt.Errorf("toCcy is required")
	}
	if len(fromCurrency) != 3 || len(toCurrency) != 3 {
		return "", "", "", fmt.Errorf("fromCcy and toCcy must be 3 characters")
	}
	if fromCurrency == toCurrency {
		return "", "", "", fmt.Errorf("fromCcy and toCcy cannot be the same")
	}

	parsedAmount, err := decimal.NewFromString(trimmedAmount)
	if err != nil {
		return "", "", "", fmt.Errorf("amount must be numeric: %w", err)
	}
	if parsedAmount.LessThanOrEqual(decimal.Zero) {
		return "", "", "", fmt.Errorf("amount must be greater than zero")
	}

	rate, err := s.rateRepo.GetRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		return "", "", "", err
	}

	usedRate, parseErr := decimal.NewFromString(strings.TrimSpace(rate.Rate))
	if parseErr != nil {
		return "", "", "", fmt.Errorf("invalid stored rate: %w", parseErr)
	}
	if usedRate.Equal(decimal.Zero) {
		return "", "", "", fmt.Errorf("rate cannot be zero")
	}

	converted := parsedAmount.Mul(usedRate)
	return converted.StringFixed(8), usedRate.StringFixed(8), rate.RateDate.Format("2006-01-02"), nil
}

func (s *RateService) GetCcyRates(ctx context.Context, req models.GetCcyRatesRequest) (models.Response[models.GetCcyRatesResponse], error) {
	logger.Info("rate service convert fcy amount request", logger.Fields{
		"payload": logger.SanitizePayload(req),
	})

	if err := req.Validate(); err != nil {
		logger.Error("rate service convert fcy amount validation failed", err, nil)
		return models.ErrorResponse[models.GetCcyRatesResponse]("validation failed", err.Error()), err
	}

	convertedAmount, rateUsed, rateDate, err := s.ConvertRate(ctx, req.Amount, req.FromCcy, req.ToCcy)
	if err != nil {
		logger.Error("rate service convert fcy amount failed", err, logger.Fields{
			"fromCcy": req.FromCcy,
			"toCcy":   req.ToCcy,
		})
		if errors.Is(err, domain.ErrRecordNotFound) {
			return models.ErrorResponse[models.GetCcyRatesResponse]("Rate not found for currency pair"), err
		}
		return models.ErrorResponse[models.GetCcyRatesResponse]("failed to get currency rates", "Unable to fetch currency rates right now"), err
	}

	response := models.GetCcyRatesResponse{
		Amount:          strings.TrimSpace(req.Amount),
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

	return models.SuccessResponse("currency rate fetched successfully", response), nil
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
