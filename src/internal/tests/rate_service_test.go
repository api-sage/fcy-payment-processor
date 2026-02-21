package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/domain"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/services"
	"github.com/shopspring/decimal"
)

type rateRepoStub struct {
	getRatesFn func(ctx context.Context) ([]domain.Rate, error)
	getRateFn  func(ctx context.Context, fromCurrency string, toCurrency string) (domain.Rate, error)
}

func (s rateRepoStub) GetRates(ctx context.Context) ([]domain.Rate, error) {
	if s.getRatesFn != nil {
		return s.getRatesFn(ctx)
	}
	return nil, nil
}

func (s rateRepoStub) GetRate(ctx context.Context, fromCurrency string, toCurrency string) (domain.Rate, error) {
	if s.getRateFn != nil {
		return s.getRateFn(ctx, fromCurrency, toCurrency)
	}
	return domain.Rate{}, nil
}

func TestRateServiceGetRatesSuccess(t *testing.T) {
	svc := services.NewRateService(rateRepoStub{
		getRatesFn: func(context.Context) ([]domain.Rate, error) {
			return []domain.Rate{
				{
					ID:           1,
					FromCurrency: "USD",
					ToCurrency:   "NGN",
					Rate:         decimal.NewFromInt(1500),
					RateDate:     time.Now().UTC(),
					CreatedAt:    time.Now().UTC(),
				},
			}, nil
		},
	})

	resp, err := svc.GetRates(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !resp.Success || resp.Data == nil || len(*resp.Data) != 1 {
		t.Fatal("expected one rate in successful response")
	}
}

func TestRateServiceGetRateSameCurrency(t *testing.T) {
	svc := services.NewRateService(nil)

	resp, err := svc.GetRate(context.Background(), models.GetRateRequest{
		FromCurrency: "USD",
		ToCurrency:   "USD",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !resp.Success || resp.Data == nil {
		t.Fatal("expected successful response with data")
	}
	if !resp.Data.Rate.Equal(decimal.NewFromInt(1)) {
		t.Fatalf("expected rate 1, got %s", resp.Data.Rate.String())
	}
}

func TestRateServiceConvertRateSameCurrency(t *testing.T) {
	svc := services.NewRateService(nil)

	converted, rateUsed, _, err := svc.ConvertRate(context.Background(), decimal.NewFromInt(200), "EUR", "EUR")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !converted.Equal(decimal.NewFromInt(200)) {
		t.Fatalf("expected converted amount 200, got %s", converted.String())
	}
	if !rateUsed.Equal(decimal.NewFromInt(1)) {
		t.Fatalf("expected rate used 1, got %s", rateUsed.String())
	}
}

func TestRateServiceGetCcyRatesSuccess(t *testing.T) {
	svc := services.NewRateService(nil)

	resp, err := svc.GetCcyRates(context.Background(), models.GetCcyRatesRequest{
		Amount:  decimal.NewFromInt(50),
		FromCcy: "GBP",
		ToCcy:   "GBP",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !resp.Success || resp.Data == nil {
		t.Fatal("expected successful response with data")
	}
	if !resp.Data.ConvertedAmount.Equal(decimal.NewFromInt(50)) {
		t.Fatalf("expected converted amount 50, got %s", resp.Data.ConvertedAmount.String())
	}
}

