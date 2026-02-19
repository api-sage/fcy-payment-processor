package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
)

type RateService struct {
	rateRepo domain.RateRepository
}

func NewRateService(rateRepo domain.RateRepository) *RateService {
	return &RateService{rateRepo: rateRepo}
}

func (s *RateService) GetRates(ctx context.Context) (models.Response[[]models.RateResponse], error) {
	rates, err := s.rateRepo.GetRates(ctx)
	if err != nil {
		return models.ErrorResponse[[]models.RateResponse]("failed to get rates", err.Error()), err
	}

	resp := make([]models.RateResponse, 0, len(rates))
	for _, rate := range rates {
		resp = append(resp, mapRateToResponse(rate))
	}

	return models.SuccessResponse("rates fetched successfully", resp), nil
}

func (s *RateService) GetRate(ctx context.Context, req models.GetRateRequest) (models.Response[models.RateResponse], error) {
	if err := req.Validate(); err != nil {
		return models.ErrorResponse[models.RateResponse]("validation failed", err.Error()), err
	}

	fromCurrency := strings.ToUpper(strings.TrimSpace(req.FromCurrency))
	toCurrency := strings.ToUpper(strings.TrimSpace(req.ToCurrency))

	rate, err := s.rateRepo.GetRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		return models.ErrorResponse[models.RateResponse]("failed to get rate", err.Error()), err
	}

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

	parsedAmount, err := strconv.ParseFloat(trimmedAmount, 64)
	if err != nil {
		return "", "", "", fmt.Errorf("amount must be numeric: %w", err)
	}
	if parsedAmount <= 0 {
		return "", "", "", fmt.Errorf("amount must be greater than zero")
	}

	rate, err := s.rateRepo.GetRate(ctx, fromCurrency, toCurrency)
	if err == nil {
		usedRate, parseErr := strconv.ParseFloat(rate.SellRate, 64)
		if parseErr != nil {
			return "", "", "", fmt.Errorf("invalid stored sell rate: %w", parseErr)
		}

		converted := parsedAmount * usedRate
		return fmt.Sprintf("%.8f", converted), fmt.Sprintf("%.8f", usedRate), rate.RateDate.Format("2006-01-02"), nil
	}

	inverseRate, inverseErr := s.rateRepo.GetRate(ctx, toCurrency, fromCurrency)
	if inverseErr != nil {
		return "", "", "", err
	}

	inverseValue, parseErr := strconv.ParseFloat(inverseRate.SellRate, 64)
	if parseErr != nil {
		return "", "", "", fmt.Errorf("invalid stored inverse sell rate: %w", parseErr)
	}
	if inverseValue == 0 {
		return "", "", "", fmt.Errorf("inverse rate cannot be zero")
	}

	usedRate := 1 / inverseValue
	converted := parsedAmount * usedRate
	return fmt.Sprintf("%.8f", converted), fmt.Sprintf("%.8f", usedRate), inverseRate.RateDate.Format("2006-01-02"), nil
}

func (s *RateService) GetCcyRates(ctx context.Context, req models.GetCcyRatesRequest) (models.Response[models.GetCcyRatesResponse], error) {
	if err := req.Validate(); err != nil {
		return models.ErrorResponse[models.GetCcyRatesResponse]("validation failed", err.Error()), err
	}

	convertedAmount, rateUsed, rateDate, err := s.ConvertRate(ctx, req.Amount, req.FromCcy, req.ToCcy)
	if err != nil {
		return models.ErrorResponse[models.GetCcyRatesResponse]("failed to get currency rates", err.Error()), err
	}

	response := models.GetCcyRatesResponse{
		Request: models.GetCcyRatesRequest{
			Amount:  strings.TrimSpace(req.Amount),
			FromCcy: strings.ToUpper(strings.TrimSpace(req.FromCcy)),
			ToCcy:   strings.ToUpper(strings.TrimSpace(req.ToCcy)),
		},
		ConvertedAmount: convertedAmount,
		RateUsed:        rateUsed,
		RateDate:        rateDate,
	}

	return models.SuccessResponse("currency rate fetched successfully", response), nil
}

func mapRateToResponse(rate domain.Rate) models.RateResponse {
	return models.RateResponse{
		ID:           rate.ID,
		FromCurrency: rate.FromCurrency,
		ToCurrency:   rate.ToCurrency,
		SellRate:     rate.SellRate,
		BuyRate:      rate.BuyRate,
		RateDate:     rate.RateDate.Format("2006-01-02"),
		CreatedAt:    rate.CreatedAt.Format(time.RFC3339),
	}
}
