package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/domain"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/services"
)

type transactionLookupRepoStub struct {
	findFn   func(ctx context.Context, reference string) ([]domain.TransientAccountTransaction, error)
	findHits int
	lastRef  string
}

func (s *transactionLookupRepoStub) Create(context.Context, domain.TransientAccountTransaction) (domain.TransientAccountTransaction, error) {
	return domain.TransientAccountTransaction{}, nil
}

func (s *transactionLookupRepoStub) List(context.Context, domain.TransientAccountTransactionFilter) ([]domain.TransientAccountTransaction, int64, error) {
	return nil, 0, nil
}

func (s *transactionLookupRepoStub) FindByExternalReference(ctx context.Context, reference string) ([]domain.TransientAccountTransaction, error) {
	s.findHits++
	s.lastRef = reference
	if s.findFn != nil {
		return s.findFn(ctx, reference)
	}
	return nil, nil
}

func TestTransactionServiceGetTransactionReturnsAllLegs(t *testing.T) {
	now := time.Now().UTC()
	repo := &transactionLookupRepoStub{
		findFn: func(context.Context, string) ([]domain.TransientAccountTransaction, error) {
			return []domain.TransientAccountTransaction{
				{ID: "leg-1", ExternalRefernece: "TXN-1", DebitedAccount: "A", EntryType: "DEBIT", CreatedAt: now},
				{ID: "leg-2", ExternalRefernece: "TXN-1", CreditedAccount: "B", EntryType: "CREDIT", CreatedAt: now},
			}, nil
		},
	}
	svc := services.NewTransactionService(repo)

	resp, err := svc.GetTransaction(context.Background(), models.GetTransactionRequest{TransactionReference: "  TXN-1  "})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !resp.Success || resp.Data == nil {
		t.Fatal("expected successful response with data")
	}
	if repo.lastRef != "TXN-1" {
		t.Fatalf("expected trimmed reference passed to repo, got %q", repo.lastRef)
	}
	if resp.Data.TransactionReference != "TXN-1" || resp.Data.Total != 2 || len(resp.Data.Items) != 2 {
		t.Fatalf("unexpected response: %+v", resp.Data)
	}
	if resp.Data.Items[0].ID != "leg-1" || resp.Data.Items[1].ID != "leg-2" {
		t.Fatalf("unexpected item order: %+v", resp.Data.Items)
	}
}

func TestTransactionServiceGetTransactionNoMatchReturnsEmptyList(t *testing.T) {
	svc := services.NewTransactionService(&transactionLookupRepoStub{})

	resp, err := svc.GetTransaction(context.Background(), models.GetTransactionRequest{TransactionReference: "missing"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp.Data == nil || resp.Data.Items == nil || len(resp.Data.Items) != 0 || resp.Data.Total != 0 {
		t.Fatalf("expected empty non-nil items, got %+v", resp.Data)
	}
}

func TestTransactionServiceGetTransactionValidation(t *testing.T) {
	tests := []struct {
		name string
		ref  string
	}{
		{name: "empty", ref: ""},
		{name: "whitespace only", ref: "   "},
		{name: "too long", ref: strings.Repeat("x", 65)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &transactionLookupRepoStub{}
			svc := services.NewTransactionService(repo)

			resp, err := svc.GetTransaction(context.Background(), models.GetTransactionRequest{TransactionReference: tt.ref})
			if err == nil {
				t.Fatal("expected validation error")
			}
			if resp.Message != "validation failed" {
				t.Fatalf("expected validation failed message, got %q", resp.Message)
			}
			if repo.findHits != 0 {
				t.Fatal("repository must not be called for invalid requests")
			}
		})
	}
}

func TestTransactionServiceGetTransactionRepoError(t *testing.T) {
	repo := &transactionLookupRepoStub{
		findFn: func(context.Context, string) ([]domain.TransientAccountTransaction, error) {
			return nil, errors.New("db down")
		},
	}
	svc := services.NewTransactionService(repo)

	resp, err := svc.GetTransaction(context.Background(), models.GetTransactionRequest{TransactionReference: "TXN-1"})
	if err == nil {
		t.Fatal("expected repository error")
	}
	if resp.Message != "failed to get transaction" {
		t.Fatalf("unexpected message %q", resp.Message)
	}
}
