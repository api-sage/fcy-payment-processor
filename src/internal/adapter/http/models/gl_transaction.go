package models

import (
	"errors"
	"math"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const (
	GLTransactionsDefaultPageSize = 20
	GLTransactionsMaxPageSize     = 100

	glTransactionDateLayout = "2006-01-02"
)

type GLTransactionResponse struct {
	ID                string          `json:"id"`
	TransferID        string          `json:"transferId"`
	ExternalReference string          `json:"externalReference"`
	DebitedAccount    string          `json:"debitedAccount"`
	CreditedAccount   string          `json:"creditedAccount"`
	EntryType         string          `json:"entryType"`
	Currency          string          `json:"currency"`
	Amount            decimal.Decimal `json:"amount"`
	CreatedAt         string          `json:"createdAt"`
}

type GetGLTransactionsRequest struct {
	StartDate       string
	EndDate         string
	GLAccountNumber string
	Page            int
	PageSize        int
}

// Validate checks the request. All filters are optional. Dates must be YYYY-MM-DD and startDate must not be after endDate.
func (r GetGLTransactionsRequest) Validate() error {
	var errs []string

	if _, _, err := r.DateRange(); err != nil {
		errs = append(errs, err.Error())
	}

	if r.Page < 1 {
		errs = append(errs, "page must be greater than zero")
	}

	pageSizeValid := r.PageSize >= 1 && r.PageSize <= GLTransactionsMaxPageSize
	if !pageSizeValid {
		errs = append(errs, "pageSize must be between 1 and 100")
	}

	if r.Page >= 1 && pageSizeValid && int64(r.Page-1) > math.MaxInt64/int64(r.PageSize) {
		errs = append(errs, "page is too large")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

// DateRange returns the parsed date bounds. The start bound is inclusive. The end bound is returned exclusive,
// as the start of the day after endDate, so that endDate covers the whole day. A nil bound means it was not provided.
func (r GetGLTransactionsRequest) DateRange() (*time.Time, *time.Time, error) {
	var start, end *time.Time

	if raw := strings.TrimSpace(r.StartDate); raw != "" {
		parsed, err := time.Parse(glTransactionDateLayout, raw)
		if err != nil {
			return nil, nil, errors.New("startDate must be in YYYY-MM-DD format")
		}
		start = &parsed
	}

	if raw := strings.TrimSpace(r.EndDate); raw != "" {
		parsed, err := time.Parse(glTransactionDateLayout, raw)
		if err != nil {
			return nil, nil, errors.New("endDate must be in YYYY-MM-DD format")
		}
		if start != nil && parsed.Before(*start) {
			return nil, nil, errors.New("startDate must not be after endDate")
		}
		exclusive := parsed.AddDate(0, 0, 1)
		end = &exclusive
	}

	return start, end, nil
}

type GetGLTransactionsResponse struct {
	Items      []GLTransactionResponse `json:"items"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"pageSize"`
	Total      int64                   `json:"total"`
	TotalPages int                     `json:"totalPages"`
}
