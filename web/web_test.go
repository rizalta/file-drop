package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSPAHandler(t *testing.T) {
	handler, err := SPAHandler()
	if err != nil {
		t.Fatalf("failed to create SPAHandler: %v", err)
	}

	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
	}{
		{
			name:           "Root path serves index",
			path:           "/",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "SPA route falls back to index",
			path:           "/f/test-id-123",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Share SPA route with query params falls back to index",
			path:           "/f/test-id-123?share=true",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Direct logo.svg static asset request",
			path:           "/logo.svg",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "HEAD request to SPA route",
			path:           "/f/test-id-123",
			method:         http.MethodHead,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Direct index.html request redirects to root",
			path:           "/index.html",
			method:         http.MethodGet,
			expectedStatus: http.StatusMovedPermanently,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := tt.method
			if method == "" {
				method = http.MethodGet
			}
			req := httptest.NewRequest(method, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d for path %s", tt.expectedStatus, rec.Code, tt.path)
			}
		})
	}
}
