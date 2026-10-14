package auth

import (
	"net/http"
	"testing"
)

func TestGetBearerToken(t *testing.T) {
	for _, tc := range []struct {
		name, header, want string
		wantError          bool
	}{
		{"valid", "Bearer token-string", "token-string", false},
		{"whitespace", "  Bearer   token-string  ", "token-string", false},
		{"case insensitive scheme", "bearer token-string", "token-string", false},
		{"missing", "", "", true},
		{"missing token", "Bearer", "", true},
		{"wrong scheme", "Basic token-string", "", true},
		{"extra field", "Bearer token extra", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			headers := http.Header{}
			if tc.header != "" {
				headers.Set("Authorization", tc.header)
			}
			got, err := GetBearerToken(headers)
			if (err != nil) != tc.wantError || got != tc.want {
				t.Fatalf("got %q, error %v; want %q, error %v", got, err, tc.want, tc.wantError)
			}
		})
	}
}

func TestDuplicateAuthorizationIsRejected(t *testing.T) {
	for _, scheme := range []string{"Bearer", "ApiKey"} {
		headers := http.Header{}
		headers.Add("Authorization", scheme+" first")
		headers.Add("Authorization", scheme+" second")
		if scheme == "Bearer" {
			if _, err := GetBearerToken(headers); err == nil {
				t.Fatal("accepted duplicate bearer headers")
			}
		} else {
			if _, err := GetAPIKey(headers); err == nil {
				t.Fatal("accepted duplicate API key headers")
			}
		}
	}
}
