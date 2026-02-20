package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/logger"
	"github.com/shopspring/decimal"
)

type ChargesService struct {
	chargePercent float64
	vatPercent    float64
}

func NewChargesService(chargePercent float64, vatPercent float64) *ChargesService {
	return &ChargesService{
		chargePercent: chargePercent,
		vatPercent:    vatPercent,
	}
}

func (s *ChargesService) GetChargesSummary(ctx context.Context, req models.GetChargesRequest) (models.Response[models.GetChargesResponse], error) {
	logger.Info("charges service get charges request", logger.Fields{
		"payload": logger.SanitizePayload(req),
	})

	_ = ctx
	if err := req.Validate(); err != nil {
		logger.Error("charges service get charges validation failed", err, nil)
		return models.ErrorResponse[models.GetChargesResponse]("validation failed", err.Error()), err
	}

	amount, currency, charge, vat, sumTotal, err := s.GetCharges(req.Amount, req.FromCurrency)
	if err != nil {
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

func (s *ChargesService) GetCharges(amount string, fromCurrency string) (string, string, string, string, string, error) {
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

	chargePercent := decimal.NewFromFloat(s.chargePercent).Div(decimal.NewFromInt(100))
	vatPercent := decimal.NewFromFloat(s.vatPercent).Div(decimal.NewFromInt(100))

	chargeValue := amountValue.Mul(chargePercent)
	vatValue := amountValue.Mul(vatPercent)
	totalValue := amountValue.Add(chargeValue).Add(vatValue)

	return amountValue.StringFixed(2), ccy, chargeValue.StringFixed(2), vatValue.StringFixed(2), totalValue.StringFixed(2), nil
}
