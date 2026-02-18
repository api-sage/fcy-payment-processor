package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/api-sage/ccy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
)

type UserService struct {
	userRepo domain.UserRepository
}

func NewUserService(userRepo domain.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) CreateUser(ctx context.Context, req models.CreateUserRequest) (models.Response[models.CreateUserResponse], error) {
	if err := req.Validate(); err != nil {
		return models.ErrorResponse[models.CreateUserResponse]("validation failed", err.Error()), err
	}

	dob, err := time.Parse("2006-01-02", strings.TrimSpace(req.DOB))
	if err != nil {
		return models.ErrorResponse[models.CreateUserResponse]("validation failed", "dob must be in YYYY-MM-DD format"), err
	}

	var middleName *string
	if trimmed := strings.TrimSpace(req.MiddleName); trimmed != "" {
		middleName = &trimmed
	}

	user := domain.User{
		CustomerID:        generateCustomerID(),
		FirstName:         strings.TrimSpace(req.FirstName),
		MiddleName:        middleName,
		LastName:          strings.TrimSpace(req.LastName),
		DOB:               dob,
		PhoneNumber:       strings.TrimSpace(req.PhoneNumber),
		IDType:            domain.IDType(strings.TrimSpace(req.IDType)),
		IDNumber:          strings.TrimSpace(req.IDNumber),
		KYCLevel:          req.KYCLevel,
		TransactionPinHas: strings.TrimSpace(req.TransactionPinHas),
	}

	created, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return models.ErrorResponse[models.CreateUserResponse]("failed to create user", err.Error()), err
	}

	response := models.CreateUserResponse{
		ID:         created.ID,
		CustomerID: created.CustomerID,
		FirstName:  created.FirstName,
		LastName:   created.LastName,
	}

	return models.SuccessResponse("user created successfully", response), nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (models.Response[models.GetUserResponse], error) {
	if strings.TrimSpace(id) == "" {
		return models.ErrorResponse[models.GetUserResponse]("validation failed", "id is required"), fmt.Errorf("id is required")
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return models.ErrorResponse[models.GetUserResponse]("failed to get user", err.Error()), err
	}

	response := models.GetUserResponse{
		ID:                user.ID,
		CustomerID:        user.CustomerID,
		FirstName:         user.FirstName,
		MiddleName:        user.MiddleName,
		LastName:          user.LastName,
		DOB:               user.DOB.Format("2006-01-02"),
		PhoneNumber:       user.PhoneNumber,
		IDType:            string(user.IDType),
		IDNumber:          user.IDNumber,
		KYCLevel:          user.KYCLevel,
		TransactionPinHas: user.TransactionPinHas,
		CreatedAt:         user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         user.UpdatedAt.Format(time.RFC3339),
	}

	return models.SuccessResponse("user fetched successfully", response), nil
}

func generateCustomerID() string {
	return fmt.Sprintf("CUST%012d", time.Now().UnixNano()%1_000_000_000_000)
}
