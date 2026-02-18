package config

import (
	"os"
	"path/filepath"
	"strings"
)

const defaultConnectionString = "Host=localhost;Port=5432;Database=payment_system_db;Username=postgres;Password=1&i355O8;Timeout=30;CommandTimeout=30"
const defaultChannelID = "GreyApp"
const defaultChannelKey = "GreyhoundKey001"
const defaultGreyBankCode = "100100"

type Config struct {
	DatabaseDSN   string
	MigrationsDir string
	ChannelID     string
	ChannelKey    string
	GreyBankCode  string
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

	return Config{
		DatabaseDSN:   normalizeConnectionString(conn),
		MigrationsDir: filepath.Join("src", "migrations"),
		ChannelID:     channelID,
		ChannelKey:    channelKey,
		GreyBankCode:  greyBankCode,
	}, nil
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
