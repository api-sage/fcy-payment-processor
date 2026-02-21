package services_test

import (
	"context"
	"testing"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/services"
	"github.com/shopspring/decimal"
)

func TestChargesServiceGetChargesSummarySuccessUSD(t *testing.T) {
	svc := services.NewChargesService(
		nil,
		decimal.NewFromInt(1),
		decimal.NewFromFloat(7.5),
		decimal.NewFromInt(2),
		decimal.NewFromInt(20),
	)

	resp, err := svc.GetChargesSummary(context.Background(), models.GetChargesRequest{
		Amount:       decimal.NewFromInt(100),
		FromCurrency: "USD",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !resp.Success || resp.Data == nil {
		t.Fatal("expected successful response with data")
	}
	if !resp.Data.Charge.Equal(decimal.NewFromInt(2)) {
		t.Fatalf("expected charge 2, got %s", resp.Data.Charge.String())
	}
	if !resp.Data.VAT.Equal(decimal.NewFromFloat(7.5)) {
		t.Fatalf("expected vat 7.5, got %s", resp.Data.VAT.String())
	}
	if !resp.Data.SumTotal.Equal(decimal.NewFromFloat(109.5)) {
		t.Fatalf("expected sumTotal 109.5, got %s", resp.Data.SumTotal.String())
	}
}

func TestChargesServiceGetChargesValidationError(t *testing.T) {
	svc := services.NewChargesService(
		nil,
		decimal.NewFromInt(1),
		decimal.NewFromFloat(7.5),
		decimal.NewFromInt(2),
		decimal.NewFromInt(20),
	)

	_, _, _, _, _, err := svc.GetCharges(context.Background(), decimal.Zero, "")
	if err == nil {
		t.Fatal("expected validation error for invalid charge inputs")
	}
}

