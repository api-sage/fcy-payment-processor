package usecase

import (
	"fmt"
	"strconv"
	"strings"
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

	amountValue, err := strconv.ParseFloat(trimmedAmount, 64)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("amount must be numeric: %w", err)
	}
	if amountValue <= 0 {
		return "", "", "", "", "", fmt.Errorf("amount must be greater than zero")
	}

	chargeValue := amountValue * (s.chargePercent / 100)
	vatValue := amountValue * (s.vatPercent / 100)
	totalValue := amountValue + chargeValue + vatValue

	return fmt.Sprintf("%.2f", amountValue), ccy, fmt.Sprintf("%.2f", chargeValue), fmt.Sprintf("%.2f", vatValue), fmt.Sprintf("%.2f", totalValue), nil
}
