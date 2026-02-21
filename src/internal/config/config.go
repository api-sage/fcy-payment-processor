package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shopspring/decimal"
)

const defaultConnectionString = "Host=localhost;Port=5432;Database=payment_system_db;Username=postgres;Password=1&i355O8;Timeout=30;CommandTimeout=30"
const defaultChannelID = "GreyApp"
const defaultChannelKey = "GreyHoundKey001"
const defaultGreyBankCode = "100100"
const defaultChargePercent = "1"
const defaultVATPercent = "7.5"
const defaultChargeMinAmount = "2"
const defaultChargeMaxAmount = "20"
const defaultInternalTransientAccountNumber = "0123456890"
const defaultInternalChargesAccountNumber = "0123445521"
const defaultInternalVATAccountNumber = "0125548976"

type Config struct {
	DatabaseDSN                    string
	MigrationsDir                  string
	ChannelID                      string
	ChannelKey                     string
	GreyBankCode                   string
	ChargePercent                  decimal.Decimal
	VATPercent                     decimal.Decimal
	ChargeMinAmount                decimal.Decimal
	ChargeMaxAmount                decimal.Decimal
	InternalTransientAccountNumber string
	InternalChargesAccountNumber   string
	InternalVATAccountNumber       string
}

func Load() (Config, error) {
	conn := strings.TrimSpace(os.Getenv("DATABASE_DSN"))
	if conn == "" {
		conn = defaultConnectionString
	}

	channelID := strings.TrimSpace(os.Getenv("CHANNEL_ID"))
	if channelID == "" {
		channelID = defaultChannelID
	}

	channelKey := strings.TrimSpace(os.Getenv("CHANNEL_KEY"))
	if channelKey == "" {
		channelKey = defaultChannelKey
	}

	greyBankCode := strings.TrimSpace(os.Getenv("GREY_BANK_CODE"))
	if greyBankCode == "" {
		greyBankCode = defaultGreyBankCode
	}

	chargePercent, err := parseDecimalEnv("CHARGE_PERCENT", defaultChargePercent)
	if err != nil {
		return Config{}, err
	}

	vatPercent, err := parseDecimalEnv("VAT_PERCENT", defaultVATPercent)
	if err != nil {
		return Config{}, err
	}

	chargeMin, err := parseDecimalEnv("CHARGE_MIN_AMOUNT", defaultChargeMinAmount)
	if err != nil {
		return Config{}, err
	}

	chargeMax, err := parseDecimalEnv("CHARGE_MAX_AMOUNT", defaultChargeMaxAmount)
	if err != nil {
		return Config{}, err
	}

	internalTransientAccountNumber := strings.TrimSpace(os.Getenv("INTERNAL_TRANSIENT_ACCOUNT_NUMBER"))
	if internalTransientAccountNumber == "" {
		internalTransientAccountNumber = defaultInternalTransientAccountNumber
	}

	internalChargesAccountNumber := strings.TrimSpace(os.Getenv("INTERNAL_CHARGES_ACCOUNT_NUMBER"))
	if internalChargesAccountNumber == "" {
		internalChargesAccountNumber = defaultInternalChargesAccountNumber
	}

	internalVATAccountNumber := strings.TrimSpace(os.Getenv("INTERNAL_VAT_ACCOUNT_NUMBER"))
	if internalVATAccountNumber == "" {
		internalVATAccountNumber = defaultInternalVATAccountNumber
	}

	return Config{
		DatabaseDSN:                    normalizeConnectionString(conn),
		MigrationsDir:                  filepath.Join("src", "migrations"),
		ChannelID:                      channelID,
		ChannelKey:                     channelKey,
		GreyBankCode:                   greyBankCode,
		ChargePercent:                  chargePercent,
		VATPercent:                     vatPercent,
		ChargeMinAmount:                chargeMin,
		ChargeMaxAmount:                chargeMax, // For now, we set max charge amount same as min to disable it
		InternalTransientAccountNumber: internalTransientAccountNumber,
		InternalChargesAccountNumber:   internalChargesAccountNumber,
		InternalVATAccountNumber:       internalVATAccountNumber,
	}, nil
}

func parseDecimalEnv(key string, fallback string) (decimal.Decimal, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		raw = fallback
	}

	value, err := decimal.NewFromString(raw)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("invalid %s: %w", key, err)
	}
	if value.LessThan(decimal.Zero) {
		return decimal.Decimal{}, fmt.Errorf("%s cannot be negative", key)
	}

	return value, nil
}

func normalizeConnectionString(raw string) string {
	parts := strings.Split(raw, ";")
	out := make([]string, 0, len(parts))
	hasSSLMode := false

	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}

		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(kv[0]))
		val := strings.TrimSpace(kv[1])

		switch key {
		case "host":
			out = append(out, "host="+val)
		case "port":
			out = append(out, "port="+val)
		case "database":
			out = append(out, "dbname="+val)
		case "username":
			out = append(out, "user="+val)
		case "password":
			out = append(out, "password="+val)
		case "timeout", "connect timeout":
			out = append(out, "connect_timeout="+val)
		case "commandtimeout", "command timeout":
			out = append(out, "statement_timeout="+val+"s")
		case "sslmode":
			hasSSLMode = true
			out = append(out, "sslmode="+val)
		default:
			out = append(out, key+"="+val)
		}
	}

	if len(out) == 0 {
		return raw
	}

	if !hasSSLMode {
		out = append(out, "sslmode=disable")
	}

	return strings.Join(out, " ")
}
