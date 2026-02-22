package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/repository/repo_interfaces"
	"github.com/api-sage/fcy-payment-processor/src/internal/commons"
	"github.com/api-sage/fcy-payment-processor/src/internal/logger"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/service_interfaces"
	"github.com/shopspring/decimal"
)

// Verify that ChargesService implements the service_interfaces.ChargesService interface
var _ service_interfaces.ChargesService = (*ChargesService)(nil)

type ChargesService struct {
	rateRepo      repo_interfaces.RateRepository
	chargePercent decimal.Decimal
	vatPercent    decimal.Decimal
	chargeMin     decimal.Decimal
	chargeMax     decimal.Decimal
}

func NewChargesService(rateRepo repo_interfaces.RateRepository, chargePercent decimal.Decimal, vatPercent decimal.Decimal, chargeMin decimal.Decimal, chargeMax decimal.Decimal) *ChargesService {
	return &ChargesService{
		rateRepo:      rateRepo,
		chargePercent: chargePercent,
		vatPercent:    vatPercent,
		chargeMin:     chargeMin,
		chargeMax:     chargeMax,
	}
}

func (s *ChargesService) GetChargesSummary(ctx context.Context, req models.GetChargesRequest) (commons.Response[models.GetChargesResponse], error) {
	logger.Info("charges service get charges request", logger.Fields{
		"payload": logger.SanitizePayload(req),
	})

	if err := req.Validate(); err != nil {
		logger.Error("charges service get charges validation failed", err, nil)
		return commons.ErrorResponse[models.GetChargesResponse]("validation failed", err.Error()), err
	}

	amount, currency, charge, vat, sumTotal, err := s.GetCharges(ctx, req.Amount, req.FromCurrency)
	if err != nil {
		if errors.Is(err, commons.ErrRecordNotFound) {
			return commons.ErrorResponse[models.GetChargesResponse]("Rate not found for currency pair"), err
		}
		logger.Error("charges service get charges calculation failed", err, nil)
		return commons.ErrorResponse[models.GetChargesResponse]("failed to get charges", "Unable to fetch charges right now"), err
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

	return commons.SuccessResponse("charges fetched successfully", response), nil
}

func (s *ChargesService) GetCharges(ctx context.Context, amount decimal.Decimal, fromCurrency string) (decimal.Decimal, string, decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	ccy := strings.ToUpper(strings.TrimSpace(fromCurrency))

	if ccy == "" {
		return decimal.Decimal{}, "", decimal.Decimal{}, decimal.Decimal{}, decimal.Decimal{}, fmt.Errorf("fromCurrency is required")
	}
	if len(ccy) != 3 {
		return decimal.Decimal{}, "", decimal.Decimal{}, decimal.Decimal{}, decimal.Decimal{}, fmt.Errorf("fromCurrency must be 3 characters")
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Decimal{}, "", decimal.Decimal{}, decimal.Decimal{}, decimal.Decimal{}, fmt.Errorf("amount must be greater than zero")
	}

	chargePercent := s.chargePercent.Div(decimal.NewFromInt(100))
	vatPercent := s.vatPercent.Div(decimal.NewFromInt(100))
	chargeMin := s.chargeMin
	chargeMax := s.chargeMax

	calculationAmount := amount
	conversionRate := decimal.NewFromInt(1)
	var err error
	if ccy != "USD" {
		currencyToUSDRate, currencyToUSDErr := s.getCurrencyToUSDRate(ctx, ccy)
		if currencyToUSDErr != nil {
			return decimal.Decimal{}, "", decimal.Decimal{}, decimal.Decimal{}, decimal.Decimal{}, currencyToUSDErr
		}

		conversionRate, err = s.getUSDToCurrencyRate(ctx, ccy)
		if err != nil {
			return decimal.Decimal{}, "", decimal.Decimal{}, decimal.Decimal{}, decimal.Decimal{}, err
		}

		calculationAmount = amount.Mul(currencyToUSDRate)
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

	totalValue := amount.Add(chargeValue).Add(vatValue)

	return amount, ccy, chargeValue, vatValue, totalValue, nil
}

func (s *ChargesService) getUSDToCurrencyRate(ctx context.Context, currency string) (decimal.Decimal, error) {
	rate, err := s.rateRepo.GetRate(ctx, "USD", currency)
	if err == nil {
		if rate.Rate.LessThanOrEqual(decimal.Zero) {
			return decimal.Decimal{}, fmt.Errorf("stored rate must be greater than zero")
		}
		return rate.Rate, nil
	}

	reverseRate, reverseErr := s.rateRepo.GetRate(ctx, currency, "USD")
	if reverseErr != nil {
		return decimal.Decimal{}, err
	}

	if reverseRate.Rate.LessThanOrEqual(decimal.Zero) {
		return decimal.Decimal{}, fmt.Errorf("stored reverse rate must be greater than zero")
	}

	return decimal.NewFromInt(1).Div(reverseRate.Rate), nil
}

func (s *ChargesService) getCurrencyToUSDRate(ctx context.Context, currency string) (decimal.Decimal, error) {
	rate, err := s.rateRepo.GetRate(ctx, currency, "USD")
	if err == nil {
		if rate.Rate.LessThanOrEqual(decimal.Zero) {
			return decimal.Decimal{}, fmt.Errorf("stored rate must be greater than zero")
		}
		return rate.Rate, nil
	}

	reverseRate, reverseErr := s.rateRepo.GetRate(ctx, "USD", currency)
	if reverseErr != nil {
		return decimal.Decimal{}, err
	}

	if reverseRate.Rate.LessThanOrEqual(decimal.Zero) {
		return decimal.Decimal{}, fmt.Errorf("stored reverse rate must be greater than zero")
	}

	return decimal.NewFromInt(1).Div(reverseRate.Rate), nil
}
