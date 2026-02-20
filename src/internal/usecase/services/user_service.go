package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/commons"
	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
	"github.com/api-sage/ccy-payment-processor/src/internal/logger"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo domain.UserRepository
}

func NewUserService(userRepo domain.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) CreateUser(ctx context.Context, req models.CreateUserRequest) (commons.Response[models.CreateUserResponse], error) {
	logger.Info("user service create user request", logger.Fields{
		"payload": logger.SanitizePayload(req),
	})

	if err := req.Validate(); err != nil {
		logger.Error("user service create user validation failed", err, nil)
		return commons.ErrorResponse[models.CreateUserResponse]("validation failed", err.Error()), err
	}

	dob, err := time.Parse("2006-01-02", strings.TrimSpace(req.DOB))
	if err != nil {
		logger.Error("user service create user invalid dob", err, nil)
		return commons.ErrorResponse[models.CreateUserResponse]("validation failed", "dob must be in YYYY-MM-DD format"), err
	}

	var middleName *string
	if trimmed := strings.TrimSpace(req.MiddleName); trimmed != "" {
		middleName = &trimmed
	}

	hashedPin, err := hashTransactionPin(strings.TrimSpace(req.TransactionPin))
	if err != nil {
		logger.Error("user service create user hash pin failed", err, nil)
		return commons.ErrorResponse[models.CreateUserResponse]("failed to create user", "failed to hash transaction pin"), err
	}

	user := domain.User{
		CustomerID:         generateCustomerID(),
		FirstName:          strings.TrimSpace(req.FirstName),
		MiddleName:         middleName,
		LastName:           strings.TrimSpace(req.LastName),
		DOB:                dob,
		PhoneNumber:        strings.TrimSpace(req.PhoneNumber),
		IDType:             domain.IDType(strings.TrimSpace(req.IDType)),
		IDNumber:           strings.TrimSpace(req.IDNumber),
		KYCLevel:           req.KYCLevel,
		TransactionPinHash: hashedPin,
	}

	created, err := s.userRepo.Create(ctx, user)
	if err != nil {
		logger.Error("user service create user repository failed", err, logger.Fields{
			"customerId": user.CustomerID,
		})
		return commons.ErrorResponse[models.CreateUserResponse]("failed to create user", "Unable to create user right now"), err
	}

	response := models.CreateUserResponse{
		ID:         created.ID,
		CustomerID: created.CustomerID,
		FirstName:  created.FirstName,
		LastName:   created.LastName,
	}

	logger.Info("user service create user success", logger.Fields{
		"userId":      response.ID,
		"customerId":  response.CustomerID,
		"firstName":   response.FirstName,
		"lastName":    response.LastName,
		"transaction": "create",
	})

	return commons.SuccessResponse("user created successfully", response), nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (commons.Response[models.GetUserResponse], error) {
	logger.Info("user service get user request", logger.Fields{
		"userId": id,
	})

	if strings.TrimSpace(id) == "" {
		return commons.ErrorResponse[models.GetUserResponse]("validation failed", "id is required"), fmt.Errorf("id is required")
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error("user service get user failed", err, logger.Fields{
			"userId": id,
		})
		if errors.Is(err, domain.ErrRecordNotFound) {
			return commons.ErrorResponse[models.GetUserResponse]("User not found"), err
		}
		return commons.ErrorResponse[models.GetUserResponse]("failed to get user", "Unable to fetch user right now"), err
	}

	response := models.GetUserResponse{
		ID:                 user.ID,
		CustomerID:         user.CustomerID,
		FirstName:          user.FirstName,
		MiddleName:         user.MiddleName,
		LastName:           user.LastName,
		DOB:                user.DOB.Format("2006-01-02"),
		PhoneNumber:        user.PhoneNumber,
		IDType:             string(user.IDType),
		IDNumber:           user.IDNumber,
		KYCLevel:           user.KYCLevel,
		TransactionPinHash: user.TransactionPinHash,
		CreatedAt:          user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          user.UpdatedAt.Format(time.RFC3339),
	}

	logger.Info("user service get user success", logger.Fields{
		"userId":     response.ID,
		"customerId": response.CustomerID,
	})

	return commons.SuccessResponse("user fetched successfully", response), nil
}

func (s *UserService) VerifyUserPin(ctx context.Context, customerID string, pin string) (commons.Response[models.VerifyUserPinResponse], error) {
	logger.Info("user service verify pin request", logger.Fields{
		"payload": logger.SanitizePayload(map[string]string{
			"customerId": customerID,
			"pin":        pin,
		}),
	})

	customerID = strings.TrimSpace(customerID)
	pin = strings.TrimSpace(pin)

	if customerID == "" {
		return commons.ErrorResponse[models.VerifyUserPinResponse]("validation failed", "customerId is required"), fmt.Errorf("customerId is required")
	}
	if pin == "" {
		return commons.ErrorResponse[models.VerifyUserPinResponse]("validation failed", "pin is required"), fmt.Errorf("pin is required")
	}

	storedPinHash, err := s.userRepo.GetTransactionPinHashByCustomerID(ctx, customerID)
	if err != nil {
		logger.Error("user service verify pin lookup failed", err, logger.Fields{
			"customerId": customerID,
		})
		if errors.Is(err, domain.ErrRecordNotFound) {
			return commons.ErrorResponse[models.VerifyUserPinResponse]("User not found"), err
		}
		return commons.ErrorResponse[models.VerifyUserPinResponse]("failed to verify pin", "Unable to verify pin right now"), err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedPinHash), []byte(pin)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			logger.Info("user service verify pin mismatch", logger.Fields{
				"customerId": customerID,
			})
			return commons.ErrorResponse[models.VerifyUserPinResponse]("invalid pin", "provided pin does not match"), fmt.Errorf("invalid pin")
		}
		wrappedErr := fmt.Errorf("verify user pin: %w", err)
		logger.Error("user service verify pin compare failed", wrappedErr, logger.Fields{
			"customerId": customerID,
		})
		return commons.ErrorResponse[models.VerifyUserPinResponse]("failed to verify pin", "Unable to verify pin right now"), wrappedErr
	}

	response := models.VerifyUserPinResponse{
		CustomerID: customerID,
		IsValidPin: true,
	}

	logger.Info("user service verify pin success", logger.Fields{
		"customerId": customerID,
		"isValidPin": true,
	})

	return commons.SuccessResponse("pin verified successfully", response), nil
}

func generateCustomerID() string {
	return fmt.Sprintf("%010d", time.Now().UnixNano()%10_000_000_000)
}

func hashTransactionPin(pin string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash transaction pin: %w", err)
	}

	return string(hashed), nil
}
