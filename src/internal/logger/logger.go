package logger

import (
	"encoding/json"
	"log"
	"strings"
)

type Fields map[string]any

var sensitiveKeys = map[string]struct{}{
	"pin":                  {},
	"transactionpin":       {},
	"transaction_pin":      {},
	"transactionpinhash":   {},
	"transaction_pin_hash": {},
}

func Info(message string, fields Fields) {
	log.Printf("INFO %s %s", message, fieldsJSON(fields))
}

func Error(message string, err error, fields Fields) {
	base := Fields{}
	for k, v := range fields {
		base[k] = v
	}
	if err != nil {
		base["error"] = err.Error()
	}

	log.Printf("ERROR %s %s", message, fieldsJSON(base))
}

func SanitizePayload(payload any) any {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "<unavailable>"
	}

	var data any
	if err := json.Unmarshal(raw, &data); err != nil {
		return "<unavailable>"
	}

	return sanitizeValue(data)
}

func fieldsJSON(fields Fields) string {
	if fields == nil {
		fields = Fields{}
	}

	sanitized := SanitizePayload(fields)
	b, err := json.Marshal(sanitized)
	if err != nil {
		return `{}`
	}

	return string(b)
}

func sanitizeValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, inner := range typed {
			if isSensitiveKey(key) {
				out[key] = "******"
				continue
			}
			out[key] = sanitizeValue(inner)
		}
		return out
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, sanitizeValue(item))
		}
		return out
	default:
		return value
	}
}

func isSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(key), "-", ""))
	_, ok := sensitiveKeys[normalized]
	return ok
}
