package models

import (
	"errors"
	"strings"
	"time"

	"github.com/api-sage/ccy-payment-processor/src/internal/domain"
)

type CreateUserRequest struct {
	FirstName      string `json:"firstName"`
	MiddleName     string `json:"middleName,omitempty"`
	LastName       string `json:"lastName"`
	DOB            string `json:"dob"`
	PhoneNumber    string `json:"phoneNumber"`
	IDType         string `json:"idType"`
	IDNumber       string `json:"idNumber"`
	KYCLevel       int    `json:"kycLevel"`
	TransactionPin string `json:"transactionPin"`
}

func (r CreateUserRequest) Validate() error {
	var errs []string

	if strings.TrimSpace(r.FirstName) == "" {
		errs = append(errs, "firstName is required")
	}
	if strings.TrimSpace(r.LastName) == "" {
		errs = append(errs, "lastName is required")
	}
	if strings.TrimSpace(r.DOB) == "" {
		errs = append(errs, "dob is required")
	} else if _, err := time.Parse("2006-01-02", strings.TrimSpace(r.DOB)); err != nil {
		errs = append(errs, "dob must be in YYYY-MM-DD format")
	}
	if strings.TrimSpace(r.PhoneNumber) == "" {
		errs = append(errs, "phoneNumber is required")
	}
	idType := strings.TrimSpace(r.IDType)
	if idType == "" {
		errs = append(errs, "idType is required")
	} else if idType != string(domain.IDTypePassport) && idType != string(domain.IDTypeDL) {
		errs = append(errs, "idType must be Passport or DL")
	}
	if strings.TrimSpace(r.IDNumber) == "" {
		errs = append(errs, "idNumber is required")
	}
	if r.KYCLevel <= 0 {
		errs = append(errs, "kycLevel must be greater than zero")
	}
	if strings.TrimSpace(r.TransactionPin) == "" {
		errs = append(errs, "transactionPin is required")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

type CreateUserResponse struct {
	ID         string `json:"id"`
	CustomerID string `json:"customerId"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
}

type GetUserResponse struct {
	ID                string  `json:"id"`
	CustomerID        string  `json:"customerId"`
	FirstName         string  `json:"firstName"`
	MiddleName        *string `json:"middleName,omitempty"`
	LastName          string  `json:"lastName"`
	DOB               string  `json:"dob"`
	PhoneNumber       string  `json:"phoneNumber"`
	IDType            string  `json:"idType"`
	IDNumber          string  `json:"idNumber"`
	KYCLevel          int     `json:"kycLevel"`
	TransactionPinHas string  `json:"transactionPinHas"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
}

type VerifyUserPinRequest struct {
	CustomerID string `json:"customerId"`
	Pin        string `json:"pin"`
}

func (r VerifyUserPinRequest) Validate() error {
	var errs []string

	if strings.TrimSpace(r.CustomerID) == "" {
		errs = append(errs, "customerId is required")
	}
	if strings.TrimSpace(r.Pin) == "" {
		errs = append(errs, "pin is required")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

type VerifyUserPinResponse struct {
	CustomerID string `json:"customerId"`
	IsValidPin bool   `json:"isValidPin"`
}
