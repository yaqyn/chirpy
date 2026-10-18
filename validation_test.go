package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSON(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		want       int
	}{
		{"valid", `{"email":"a@example.com"}`, 200}, {"null", `null`, 400},
		{"trailing", `{} {}`, 400}, {"malformed", `{`, 400}, {"wrong type", `[]`, 400},
		{"oversized", `{"email":"` + strings.Repeat("a", maxBodyBytes) + `"}`, 413},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/", strings.NewReader(tc.body))
			r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
			var payload struct {
				Email string `json:"email"`
			}
			got := decodeJSON(w, r, &payload)
			if got != (tc.want == 200) || w.Code != tc.want {
				t.Fatalf("valid=%v status=%d", got, w.Code)
			}
		})
	}
}
func TestQueryValidation(t *testing.T) {
	for _, query := range []string{"?limit=%ZZ", "?offset=%", "?limit=1;offset=100", "?limit=5&limit=10", "?sort=asc&sort=desc"} {
		if _, err := parseChirpQuery(httptest.NewRequest("GET", "/"+query, nil)); err == nil {
			t.Fatalf("accepted malformed query %s", query)
		}
	}
	for _, query := range []string{"?sort=wrong", "?limit=0", "?limit=101", "?offset=-1", "?author_id=wrong"} {
		if _, err := parseChirpQuery(httptest.NewRequest("GET", "/"+query, nil)); err == nil {
			t.Fatalf("accepted %s", query)
		}
	}
	p, err := parseChirpQuery(httptest.NewRequest("GET", "/?sort=desc&limit=5&offset=10", nil))
	if err != nil || !p.Descending || p.PageLimit != 5 || p.PageOffset != 10 {
		t.Fatalf("params=%+v err=%v", p, err)
	}
}
func TestCleanChirp(t *testing.T) {
	for _, separator := range []string{"\t", "\n", "\r\n", "\u00a0", "\u2003"} {
		if got := cleanChirp("hello" + separator + "FORNAX" + separator); got != "hello"+separator+"****"+separator {
			t.Errorf("whitespace cleaning: %q", got)
		}
	}
	for input, want := range map[string]string{"KERFUFFLE sharbert fornax": "**** **** ****", "Sharbert!": "Sharbert!", "a  kerfuffle": "a  ****", "plain text": "plain text"} {
		if got := cleanChirp(input); got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	}
}
func TestValidEmail(t *testing.T) {
	for _, value := range []string{"", "invalid", "Name <a@example.com>"} {
		if validEmail(value) {
			t.Fatalf("accepted %q", value)
		}
	}
	if !validEmail("a@example.com") {
		t.Fatal("rejected email")
	}
}

func TestInvalidLoginInputIsRejectedWithoutDatabaseAccess(t *testing.T) {
	cfg := &apiConfig{}
	for _, body := range []string{
		`{"email":"invalid","password":"password"}`,
		`{"email":"a@example.com","password":""}`,
		`{"email":"a@example.com","password":"` + strings.Repeat("a", 1025) + `"}`,
	} {
		response := httptest.NewRecorder()
		cfg.routes().ServeHTTP(response, httptest.NewRequest("POST", "/api/login", strings.NewReader(body)))
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("got %d: %s", response.Code, response.Body)
		}
		if !strings.Contains(response.Body.String(), "Incorrect email or password") {
			t.Fatal(response.Body.String())
		}
	}
}
