package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

const defaultConnectionString = "postgresql://neondb_owner:npg_Xe6qn2RiFGpZ@ep-quiet-art-b4gedixw-pooler.c-6.us-east-2.aws.neon.tech/payment_system_db?sslmode=require&channel_binding=require"
const defaultChannelID = "CalipaAgent"
const defaultChannelKey = "CalipaKey001"
const defaultGreyBankCode = "100100"
const defaultChargePercent = "1"
const defaultVATPercent = "7.5"
const defaultChargeMinAmount = "2"
const defaultChargeMaxAmount = "20"
const defaultInternalTransientAccountNumber = "0123456890"
const defaultInternalChargesAccountNumber = "0123445521"
const defaultInternalVATAccountNumber = "0125548976"
const defaultExternalUSDGLAccountNumber = "0125548977"
const defaultExternalGBPGLAccountNumber = "0125548978"
const defaultExternalEURGLAccountNumber = "0125548979"
const defaultExternalNGNGLAccountNumber = "0125548980"
const defaultDBMaxOpenConns = "30"
const defaultDBMaxIdleConns = "20"

type Config struct {
	DatabaseDSN                    string
	DBMaxOpenConns                 int
	DBMaxIdleConns                 int
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
	ExternalUSDGLAccountNumber     string
	ExternalGBPGLAccountNumber     string
	ExternalEURGLAccountNumber     string
	ExternalNGNGLAccountNumber     string
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

	dbMaxOpenConns, err := parseIntEnv("DB_MAX_OPEN_CONNS", defaultDBMaxOpenConns)
	if err != nil {
		return Config{}, err
	}

	dbMaxIdleConns, err := parseIntEnv("DB_MAX_IDLE_CONNS", defaultDBMaxIdleConns)
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

	externalUSDGLAccountNumber := strings.TrimSpace(os.Getenv("EXTERNAL_USD_GL_ACCOUNT_NUMBER"))
	if externalUSDGLAccountNumber == "" {
		externalUSDGLAccountNumber = defaultExternalUSDGLAccountNumber
	}

	externalGBPGLAccountNumber := strings.TrimSpace(os.Getenv("EXTERNAL_GBP_GL_ACCOUNT_NUMBER"))
	if externalGBPGLAccountNumber == "" {
		externalGBPGLAccountNumber = defaultExternalGBPGLAccountNumber
	}

	externalEURGLAccountNumber := strings.TrimSpace(os.Getenv("EXTERNAL_EUR_GL_ACCOUNT_NUMBER"))
	if externalEURGLAccountNumber == "" {
		externalEURGLAccountNumber = defaultExternalEURGLAccountNumber
	}

	externalNGNGLAccountNumber := strings.TrimSpace(os.Getenv("EXTERNAL_NGN_GL_ACCOUNT_NUMBER"))
	if externalNGNGLAccountNumber == "" {
		externalNGNGLAccountNumber = defaultExternalNGNGLAccountNumber
	}

	return Config{
		DatabaseDSN:                    normalizeConnectionString(conn),
		DBMaxOpenConns:                 dbMaxOpenConns,
		DBMaxIdleConns:                 dbMaxIdleConns,
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
		ExternalUSDGLAccountNumber:     externalUSDGLAccountNumber,
		ExternalGBPGLAccountNumber:     externalGBPGLAccountNumber,
		ExternalEURGLAccountNumber:     externalEURGLAccountNumber,
		ExternalNGNGLAccountNumber:     externalNGNGLAccountNumber,
	}, nil
}

func parseIntEnv(key string, fallback string) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		raw = fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	if value < 1 {
		return 0, fmt.Errorf("%s must be at least 1", key)
	}

	return value, nil
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
	// URL-form DSNs are already valid for lib/pq; rewriting them would lowercase the password.
	if strings.HasPrefix(raw, "postgres://") || strings.HasPrefix(raw, "postgresql://") {
		return raw
	}

	// URL-form DSNs are already valid for lib/pq; rewriting them would lowercase the password.
	if strings.HasPrefix(raw, "postgres://") || strings.HasPrefix(raw, "postgresql://") {
		return raw
	}

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
