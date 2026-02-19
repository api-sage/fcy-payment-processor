package usecase

import (
	"context"
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
