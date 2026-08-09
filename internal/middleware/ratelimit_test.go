package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestUploadRateLimiter(t *testing.T) {
	r := chi.NewRouter()
	r.Use(UploadRateLimiter(2))

	r.Post("/api/upload", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/api/upload", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected request %d to be status 200, got %d", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest("POST", "/api/upload", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 3rd request to be status 429, got %d", rec.Code)
	}
}
