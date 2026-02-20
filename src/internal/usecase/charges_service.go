package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
	"github.com/api-sage/ccy-payment-processor/src/internal/logger"
	"github.com/shopspring/decimal"
)

type ChargesService struct {
	rateRepo      domain.RateRepository
	chargePercent decimal.Decimal
	vatPercent    decimal.Decimal
	chargeMin     decimal.Decimal
	chargeMax     decimal.Decimal
}

func NewChargesService(rateRepo domain.RateRepository, chargePercent decimal.Decimal, vatPercent decimal.Decimal, chargeMin decimal.Decimal, chargeMax decimal.Decimal) *ChargesService {
	return &ChargesService{
		rateRepo:      rateRepo,
		chargePercent: chargePercent,
		vatPercent:    vatPercent,
		chargeMin:     chargeMin,
		chargeMax:     chargeMax,
	}
}

func (s *ChargesService) GetChargesSummary(ctx context.Context, req models.GetChargesRequest) (models.Response[models.GetChargesResponse], error) {
	logger.Info("charges service get charges request", logger.Fields{
		"payload": logger.SanitizePayload(req),
	})

	if err := req.Validate(); err != nil {
		logger.Error("charges service get charges validation failed", err, nil)
		return models.ErrorResponse[models.GetChargesResponse]("validation failed", err.Error()), err
	}

	amount, currency, charge, vat, sumTotal, err := s.GetCharges(ctx, req.Amount, req.FromCurrency)
	if err != nil {
		if errors.Is(err, domain.ErrRecordNotFound) {
			return models.ErrorResponse[models.GetChargesResponse]("Rate not found for currency pair"), err
		}
		logger.Error("charges service get charges calculation failed", err, nil)
		return models.ErrorResponse[models.GetChargesResponse]("failed to get charges", "Unable to fetch charges right now"), err
	}

	response := models.GetChargesResponse{
		Amount:   amount,
		Currency: currency,
		Charge:   charge,
		VAT:      vat,
		SumTotal: sumTotal,
	}

	logger.Info("charges service get charges success", logger.Fields{
		"amount":   response.Amount,
		"currency": response.Currency,
		"charge":   response.Charge,
		"vat":      response.VAT,
		"sumTotal": response.SumTotal,
	})

	return models.SuccessResponse("charges fetched successfully", response), nil
}

func (s *ChargesService) GetCharges(ctx context.Context, amount string, fromCurrency string) (string, string, string, string, string, error) {
	trimmedAmount := strings.TrimSpace(amount)
	ccy := strings.ToUpper(strings.TrimSpace(fromCurrency))

	if trimmedAmount == "" {
		return "", "", "", "", "", fmt.Errorf("amount is required")
	}
	if ccy == "" {
		return "", "", "", "", "", fmt.Errorf("fromCurrency is required")
	}
	if len(ccy) != 3 {
		return "", "", "", "", "", fmt.Errorf("fromCurrency must be 3 characters")
	}

	amountValue, err := decimal.NewFromString(trimmedAmount)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("amount must be numeric: %w", err)
	}
	if amountValue.LessThanOrEqual(decimal.Zero) {
		return "", "", "", "", "", fmt.Errorf("amount must be greater than zero")
	}

	chargePercent := s.chargePercent.Div(decimal.NewFromInt(100))
	vatPercent := s.vatPercent.Div(decimal.NewFromInt(100))
	chargeMin := s.chargeMin
	chargeMax := s.chargeMax

	calculationAmount := amountValue
	conversionRate := decimal.NewFromInt(1)
	if ccy != "USD" {
		currencyToUSDRate, currencyToUSDErr := s.getCurrencyToUSDRate(ctx, ccy)
		if currencyToUSDErr != nil {
			return "", "", "", "", "", currencyToUSDErr
		}

		conversionRate, err = s.getUSDToCurrencyRate(ctx, ccy)
		if err != nil {
			return "", "", "", "", "", err
		}

		calculationAmount = amountValue.Mul(currencyToUSDRate)
	}

	chargeValue := calculationAmount.Mul(chargePercent)
	if chargeValue.LessThan(chargeMin) {
		chargeValue = chargeMin
	}
	if chargeValue.GreaterThan(chargeMax) {
		chargeValue = chargeMax
	}
	vatValue := calculationAmount.Mul(vatPercent)

	if ccy != "USD" {
		chargeValue = chargeValue.Mul(conversionRate)
		vatValue = vatValue.Mul(conversionRate)
	}

	totalValue := amountValue.Add(chargeValue).Add(vatValue)

	return amountValue.StringFixed(2), ccy, chargeValue.StringFixed(2), vatValue.StringFixed(2), totalValue.StringFixed(2), nil
}

func (s *ChargesService) getUSDToCurrencyRate(ctx context.Context, currency string) (decimal.Decimal, error) {
	rate, err := s.rateRepo.GetRate(ctx, "USD", currency)
	if err == nil {
		parsed, parseErr := decimal.NewFromString(strings.TrimSpace(rate.Rate))
		if parseErr != nil {
			return decimal.Decimal{}, fmt.Errorf("invalid stored rate: %w", parseErr)
		}
		if parsed.LessThanOrEqual(decimal.Zero) {
			return decimal.Decimal{}, fmt.Errorf("stored rate must be greater than zero")
		}
		return parsed, nil
	}

	reverseRate, reverseErr := s.rateRepo.GetRate(ctx, currency, "USD")
	if reverseErr != nil {
		return decimal.Decimal{}, err
	}

	parsedReverse, parseErr := decimal.NewFromString(strings.TrimSpace(reverseRate.Rate))
	if parseErr != nil {
		return decimal.Decimal{}, fmt.Errorf("invalid stored reverse rate: %w", parseErr)
	}
	if parsedReverse.LessThanOrEqual(decimal.Zero) {
		return decimal.Decimal{}, fmt.Errorf("stored reverse rate must be greater than zero")
	}

	return decimal.NewFromInt(1).Div(parsedReverse), nil
}

func (s *ChargesService) getCurrencyToUSDRate(ctx context.Context, currency string) (decimal.Decimal, error) {
	rate, err := s.rateRepo.GetRate(ctx, currency, "USD")
	if err == nil {
		parsed, parseErr := decimal.NewFromString(strings.TrimSpace(rate.Rate))
		if parseErr != nil {
			return decimal.Decimal{}, fmt.Errorf("invalid stored rate: %w", parseErr)
		}
		if parsed.LessThanOrEqual(decimal.Zero) {
			return decimal.Decimal{}, fmt.Errorf("stored rate must be greater than zero")
		}
		return parsed, nil
	}

	reverseRate, reverseErr := s.rateRepo.GetRate(ctx, "USD", currency)
	if reverseErr != nil {
		return decimal.Decimal{}, err
	}

	parsedReverse, parseErr := decimal.NewFromString(strings.TrimSpace(reverseRate.Rate))
	if parseErr != nil {
		return decimal.Decimal{}, fmt.Errorf("invalid stored reverse rate: %w", parseErr)
	}
	if parsedReverse.LessThanOrEqual(decimal.Zero) {
		return decimal.Decimal{}, fmt.Errorf("stored reverse rate must be greater than zero")
	}

	return decimal.NewFromInt(1).Div(parsedReverse), nil
}
