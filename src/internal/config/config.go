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
const defaultChargePercent = "1.0"
const defaultVATPercent = "7.5"
const defaultChargeMinAmount = "2.0"
const defaultChargeMaxAmount = "20.0"

type Config struct {
	DatabaseDSN   string
	MigrationsDir string
	ChannelID     string
	ChannelKey    string
	GreyBankCode  string
	ChargePercent decimal.Decimal
	VATPercent    decimal.Decimal
	ChargeMin     decimal.Decimal
	ChargeMax     decimal.Decimal
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
	if chargeMax.LessThan(chargeMin) {
		return Config{}, fmt.Errorf("CHARGE_MAX_AMOUNT cannot be less than CHARGE_MIN_AMOUNT")
	}

	return Config{
		DatabaseDSN:   normalizeConnectionString(conn),
		MigrationsDir: filepath.Join("src", "migrations"),
		ChannelID:     channelID,
		ChannelKey:    channelKey,
		GreyBankCode:  greyBankCode,
		ChargePercent: chargePercent,
		VATPercent:    vatPercent,
		ChargeMin:     chargeMin,
		ChargeMax:     chargeMax,
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
