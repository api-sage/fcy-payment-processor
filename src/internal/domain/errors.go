package domain

import "errors"

var ErrRecordNotFound = errors.New("Record not found")
var ErrInsufficientBalance = errors.New("Insufficient balance")
