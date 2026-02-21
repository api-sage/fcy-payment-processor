package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/domain"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/services"
	"golang.org/x/crypto/bcrypt"
)

type userRepoStub struct {
	createFn                      func(ctx context.Context, user domain.User) (domain.User, error)
	getByIDFn                     func(ctx context.Context, id string) (domain.User, error)
	getByCustomerIDFn             func(ctx context.Context, customerID string) (domain.User, error)
	updateFn                      func(ctx context.Context, user domain.User) (domain.User, error)
	getTransactionPinHashByCustFn func(ctx context.Context, customerID string) (string, error)
}

func (s userRepoStub) Create(ctx context.Context, user domain.User) (domain.User, error) {
	if s.createFn != nil {
		return s.createFn(ctx, user)
	}
	return domain.User{}, nil
}

func (s userRepoStub) GetByID(ctx context.Context, id string) (domain.User, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}
	return domain.User{}, nil
}

func (s userRepoStub) GetByCustomerID(ctx context.Context, customerID string) (domain.User, error) {
	if s.getByCustomerIDFn != nil {
		return s.getByCustomerIDFn(ctx, customerID)
	}
	return domain.User{}, nil
}

func (s userRepoStub) Update(ctx context.Context, user domain.User) (domain.User, error) {
	if s.updateFn != nil {
		return s.updateFn(ctx, user)
	}
	return domain.User{}, nil
}

func (s userRepoStub) GetTransactionPinHashByCustomerID(ctx context.Context, customerID string) (string, error) {
	if s.getTransactionPinHashByCustFn != nil {
		return s.getTransactionPinHashByCustFn(ctx, customerID)
	}
	return "", nil
}

func TestUserServiceCreateUserSuccess(t *testing.T) {
	svc := services.NewUserService(userRepoStub{
		createFn: func(_ context.Context, user domain.User) (domain.User, error) {
			if user.TransactionPinHash == "" || user.TransactionPinHash == "1234" {
				t.Fatal("expected hashed transaction pin before persistence")
			}
			user.ID = "u-1"
			user.CreatedAt = time.Now().UTC()
			user.UpdatedAt = time.Now().UTC()
			return user, nil
		},
	})

	resp, err := svc.CreateUser(context.Background(), models.CreateUserRequest{
		FirstName:      "Ada",
		LastName:       "Lovelace",
		DOB:            "1990-01-01",
		PhoneNumber:    "08010000000",
		IDType:         "Passport",
		IDNumber:       "A123456",
		KYCLevel:       1,
		TransactionPin: "1234",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !resp.Success || resp.Data == nil {
		t.Fatal("expected successful response with data")
	}
}

func TestUserServiceGetUserValidationError(t *testing.T) {
	svc := services.NewUserService(userRepoStub{})

	_, err := svc.GetUser(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error for missing user id")
	}
}

func TestUserServiceVerifyUserPinSuccess(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("4321"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to generate test hash: %v", err)
	}

	svc := services.NewUserService(userRepoStub{
		getTransactionPinHashByCustFn: func(context.Context, string) (string, error) {
			return string(hash), nil
		},
	})

	resp, verifyErr := svc.VerifyUserPin(context.Background(), "1000000001", "4321")
	if verifyErr != nil {
		t.Fatalf("expected nil error, got %v", verifyErr)
	}
	if !resp.Success || resp.Data == nil || !resp.Data.IsValidPin {
		t.Fatal("expected successful pin verification")
	}
}

