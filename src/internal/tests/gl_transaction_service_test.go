package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/domain"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/services"
)

type glTransactionRepoStub struct {
	listFn   func(ctx context.Context, filter domain.TransientAccountTransactionFilter) ([]domain.TransientAccountTransaction, int64, error)
	listHits int
}

func (s *glTransactionRepoStub) Create(context.Context, domain.TransientAccountTransaction) (domain.TransientAccountTransaction, error) {
	return domain.TransientAccountTransaction{}, nil
}

func (s *glTransactionRepoStub) List(ctx context.Context, filter domain.TransientAccountTransactionFilter) ([]domain.TransientAccountTransaction, int64, error) {
	s.listHits++
	if s.listFn != nil {
		return s.listFn(ctx, filter)
	}
	return nil, 0, nil
}

func (s *glTransactionRepoStub) FindByExternalReference(context.Context, string) ([]domain.TransientAccountTransaction, error) {
	return nil, nil
}

func TestGLTransactionServiceGetGLTransactionsPagination(t *testing.T) {
	var captured domain.TransientAccountTransactionFilter
	repo := &glTransactionRepoStub{
		listFn: func(_ context.Context, filter domain.TransientAccountTransactionFilter) ([]domain.TransientAccountTransaction, int64, error) {
			captured = filter
			return []domain.TransientAccountTransaction{{ID: "tx-1", CreatedAt: time.Now().UTC()}}, 45, nil
		},
	}
	svc := services.NewGLTransactionService(repo)

	resp, err := svc.GetGLTransactions(context.Background(), models.GetGLTransactionsRequest{
		StartDate:       "2026-01-01",
		EndDate:         "2026-01-31",
		GLAccountNumber: " 0125548977 ",
		Page:            2,
		PageSize:        20,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !resp.Success || resp.Data == nil {
		t.Fatal("expected successful response with data")
	}

	data := resp.Data
	if data.Total != 45 || data.TotalPages != 3 || data.Page != 2 || data.PageSize != 20 {
		t.Fatalf("unexpected pagination: %+v", data)
	}
	if len(data.Items) != 1 || data.Items[0].ID != "tx-1" {
		t.Fatalf("unexpected items: %+v", data.Items)
	}

	if captured.Limit != 20 || captured.Offset != 20 {
		t.Fatalf("expected limit 20 and offset 20, got limit %d offset %d", captured.Limit, captured.Offset)
	}
	if captured.GLAccountNumber != "0125548977" {
		t.Fatalf("expected trimmed account number, got %q", captured.GLAccountNumber)
	}
	if captured.StartDate == nil || !captured.StartDate.Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected start date: %v", captured.StartDate)
	}
	if captured.EndDate == nil || !captured.EndDate.Equal(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected end date to be exclusive next day 2026-02-01, got %v", captured.EndDate)
	}
}

func TestGLTransactionServiceGetGLTransactionsEmptyItemsIsNotNil(t *testing.T) {
	svc := services.NewGLTransactionService(&glTransactionRepoStub{})

	resp, err := svc.GetGLTransactions(context.Background(), models.GetGLTransactionsRequest{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp.Data == nil || resp.Data.Items == nil {
		t.Fatal("expected non-nil items slice")
	}
	if resp.Data.TotalPages != 0 {
		t.Fatalf("expected zero total pages, got %d", resp.Data.TotalPages)
	}
}

func TestGLTransactionServiceGetGLTransactionsValidation(t *testing.T) {
	tests := []struct {
		name string
		req  models.GetGLTransactionsRequest
	}{
		{name: "invalid start date", req: models.GetGLTransactionsRequest{StartDate: "01-01-2026", Page: 1, PageSize: 20}},
		{name: "invalid end date", req: models.GetGLTransactionsRequest{EndDate: "2026/01/31", Page: 1, PageSize: 20}},
		{name: "start after end", req: models.GetGLTransactionsRequest{StartDate: "2026-02-01", EndDate: "2026-01-31", Page: 1, PageSize: 20}},
		{name: "page zero", req: models.GetGLTransactionsRequest{Page: 0, PageSize: 20}},
		{name: "page size zero", req: models.GetGLTransactionsRequest{Page: 1, PageSize: 0}},
		{name: "page size over max", req: models.GetGLTransactionsRequest{Page: 1, PageSize: 101}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &glTransactionRepoStub{}
			svc := services.NewGLTransactionService(repo)

			resp, err := svc.GetGLTransactions(context.Background(), tt.req)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if resp.Message != "validation failed" {
				t.Fatalf("expected validation failed message, got %q", resp.Message)
			}
			if repo.listHits != 0 {
				t.Fatal("repository must not be called for invalid requests")
			}
		})
	}
}

func TestGLTransactionServiceGetGLTransactionsRepoError(t *testing.T) {
	repo := &glTransactionRepoStub{
		listFn: func(context.Context, domain.TransientAccountTransactionFilter) ([]domain.TransientAccountTransaction, int64, error) {
			return nil, 0, errors.New("db down")
		},
	}
	svc := services.NewGLTransactionService(repo)

	resp, err := svc.GetGLTransactions(context.Background(), models.GetGLTransactionsRequest{Page: 1, PageSize: 20})
	if err == nil {
		t.Fatal("expected repository error")
	}
	if resp.Message != "failed to get gl transactions" {
		t.Fatalf("unexpected message %q", resp.Message)
	}
}
