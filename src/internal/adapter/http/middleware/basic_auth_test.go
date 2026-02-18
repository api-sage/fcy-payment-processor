package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasicAuth_AllowsValidCredentials(t *testing.T) {
	mw := BasicAuth("GreyApp", "GrehoundKey001")
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Channel-ID", "GreyApp")
	req.Header.Set("X-Channel-Key", "GrehoundKey001")

	rr := httptest.NewRecorder()
	mw(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestBasicAuth_RejectsInvalidCredentials(t *testing.T) {
	mw := BasicAuth("GreyApp", "GrehoundKey001")
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Channel-ID", "GreyApp")
	req.Header.Set("X-Channel-Key", "WrongKey")

	rr := httptest.NewRecorder()
	mw(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}
