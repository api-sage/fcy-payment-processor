package domain

import "time"

type IDType string

const (
	IDTypePassport IDType = "Passport"
	IDTypeDL       IDType = "DL"
)

type User struct {
	ID                 string
	CustomerID         string
	FirstName          string
	MiddleName         *string
	LastName           string
	DOB                time.Time
	PhoneNumber        string
	IDType             IDType
	IDNumber           string
	KYCLevel           int
	TransactionPinHash string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
