package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/yaqyn/chirpy/internal/database"
)

// Requires a migrated disposable database; never resets shared data.
func TestAPIIntegration(t *testing.T) {
	url := os.Getenv("TEST_DB_URL")
	if url == "" {
		t.Skip("set TEST_DB_URL to run PostgreSQL integration checks")
	}
	db, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	cfg := &apiConfig{dbQueries: database.New(db), jwtSecret: strings.Repeat("x", 32), polkaKey: "integration-key"}
	router := cfg.routes()
	request := func(method, path string, payload any, authorization string, want int) map[string]any {
		t.Helper()
		var body bytes.Buffer
		if payload != nil {
			if err := json.NewEncoder(&body).Encode(payload); err != nil {
				t.Fatal(err)
			}
		}
		r := httptest.NewRequest(method, path, &body)
		if authorization != "" {
			r.Header.Set("Authorization", authorization)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code != want {
			t.Fatalf("%s %s: got %d want %d: %s", method, path, w.Code, want, w.Body.String())
		}
		if want == 204 {
			if w.Body.Len() != 0 {
				t.Fatal("204 has body")
			}
			return nil
		}
		var result map[string]any
		if strings.HasPrefix(w.Body.String(), "{") {
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
		}
		return result
	}
	email := "integration-" + uuid.NewString() + "@example.com"
	credentials := map[string]string{"email": email, "password": "test-password"}
	user := request("POST", "/api/users", credentials, "", 201)
	id := user["id"].(string)
	defer db.Exec("DELETE FROM users WHERE id=$1", id)
	if user["is_chirpy_red"] != false || user["hashed_password"] != nil {
		t.Fatal("unsafe user response")
	}
	request("POST", "/api/users", credentials, "", 409)
	request("POST", "/api/login", map[string]string{"email": email, "password": "wrong"}, "", 401)
	login := request("POST", "/api/login", credentials, "", 200)
	access := "Bearer " + login["token"].(string)
	refresh := "Bearer " + login["refresh_token"].(string)
	chirp := request("POST", "/api/chirps", map[string]string{"body": "hello KERFUFFLE"}, access, 201)
	if chirp["user_id"] != id || chirp["body"] != "hello ****" {
		t.Fatal("chirp author or cleaning wrong")
	}
	path := "/api/chirps/" + chirp["id"].(string)
	request("POST", "/api/chirps", map[string]string{"body": strings.Repeat("界", 141)}, access, 400)
	request("POST", "/api/chirps", map[string]string{"body": " "}, access, 400)
	request("POST", "/api/polka/webhooks", map[string]any{"event": "user.upgraded", "data": map[string]string{"user_id": id}}, "ApiKey wrong", 401)
	request("POST", "/api/polka/webhooks", map[string]any{"event": "user.upgraded", "data": map[string]string{"user_id": id}}, "ApiKey integration-key", 204)
	upgraded := request("POST", "/api/login", credentials, "", 200)
	if upgraded["is_chirpy_red"] != true {
		t.Fatal("upgrade not persisted")
	}
	request("POST", "/api/refresh", nil, refresh, 200)
	request("POST", "/api/revoke", nil, refresh, 204)
	request("POST", "/api/refresh", nil, refresh, 401)
	request("DELETE", path, "", access, 204)
	request("GET", path, nil, "", 404)
	updated := request("PUT", "/api/users", map[string]string{"email": email, "password": "changed-password"}, access, 200)
	if updated["id"] != id || updated["is_chirpy_red"] != true {
		t.Fatal("update lost identity/membership")
	}
	request("POST", "/api/login", credentials, "", 401)
	request("POST", "/api/login", map[string]string{"email": email, "password": "changed-password"}, "", 200)
	// List parameters are tested against real rows and preserve array shape.
	for _, query := range []string{"?sort=asc&limit=1", "?sort=desc&author_id=" + id, "?offset=100000"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/chirps"+query, nil))
		if w.Code != 200 || !strings.HasPrefix(w.Body.String(), "[") {
			t.Fatalf("list response %d %s", w.Code, w.Body)
		}
	}
}
