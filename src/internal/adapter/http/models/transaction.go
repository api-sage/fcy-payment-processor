package models

import (
	"errors"
	"strings"
)

// transactionReferenceMaxLength matches external_refernece VARCHAR(64) in transient_account_transactions.
const transactionReferenceMaxLength = 64

type GetTransactionRequest struct {
	TransactionReference string
}

// Validate checks that a transaction reference was supplied and fits the external_refernece column.
func (r GetTransactionRequest) Validate() error {
	reference := strings.TrimSpace(r.TransactionReference)
	if reference == "" {
		return errors.New("transactionReference is required")
	}
	if len(reference) > transactionReferenceMaxLength {
		return errors.New("transactionReference must not exceed 64 characters")
	}
	return nil
}

type GetTransactionResponse struct {
	TransactionReference string                  `json:"transactionReference"`
	Items                []GLTransactionResponse `json:"items"`
	Total                int                     `json:"total"`
}
