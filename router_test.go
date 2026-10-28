package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublicDirectoryIsolation(t *testing.T) {
	cfg := &apiConfig{}
	for _, path := range []string{"/app/.env", "/app/main.go", "/app/go.mod", "/app/../.env"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			cfg.routes().ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code == http.StatusOK {
				t.Fatalf("private path served: %s", path)
			}
		})
	}
	response := httptest.NewRecorder()
	cfg.routes().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/app/", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("public page: %d", response.Code)
	}
	if response.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("missing security header")
	}
	if response.Header().Get("Permissions-Policy") == "" || response.Header().Get("Cross-Origin-Resource-Policy") != "same-origin" {
		t.Fatal("missing browser isolation headers")
	}
}

func TestAuthenticationGuards(t *testing.T) {
	cfg := &apiConfig{jwtSecret: "test-secret", polkaKey: "test-key"}
	for _, endpoint := range []struct{ method, path string }{
		{"POST", "/api/chirps"}, {"DELETE", "/api/chirps/" + "00000000-0000-0000-0000-000000000001"},
		{"PUT", "/api/users"}, {"POST", "/api/refresh"}, {"POST", "/api/revoke"}, {"POST", "/api/polka/webhooks"},
	} {
		t.Run(endpoint.method+endpoint.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			cfg.routes().ServeHTTP(response, httptest.NewRequest(endpoint.method, endpoint.path, nil))
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("got %d", response.Code)
			}
		})
	}
	response := httptest.NewRecorder()
	cfg.routes().ServeHTTP(response, httptest.NewRequest("POST", "/admin/reset", nil))
	if response.Code != http.StatusForbidden {
		t.Fatalf("reset outside dev: %d", response.Code)
	}
}

func TestJSONResponsesDeclareUTF8(t *testing.T) {
	response := httptest.NewRecorder()
	respondWithJSON(response, http.StatusOK, map[string]string{"message": "hello"})
	if response.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("content type: %q", response.Header().Get("Content-Type"))
	}
}
