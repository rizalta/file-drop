package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeader(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := SecurityHeader(nextHandler)
	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()

	handlerToTest.ServeHTTP(rec, req)

	headers := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"X-XSS-Protection":        "0",
		"Referrer-Policy":         "strict-origin-when-cross-origin",
		"Content-Security-Policy": "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:",
	}

	for header, expected := range headers {
		got := rec.Header().Get(header)
		if got != expected {
			t.Errorf("expected header %s to be %q, got %q", header, expected, got)
		}
	}
}
