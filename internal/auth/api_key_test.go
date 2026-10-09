package auth

import (
	"net/http"
	"testing"
)

func TestGetAPIKey(t *testing.T) {
	for _, tc := range []struct {
		name, header, want string
		wantError          bool
	}{
		{"valid", "ApiKey test-key", "test-key", false},
		{"whitespace", "  ApiKey   test-key  ", "test-key", false},
		{"case insensitive scheme", "apikey test-key", "test-key", false},
		{"missing", "", "", true},
		{"missing key", "ApiKey", "", true},
		{"wrong scheme", "Bearer test-key", "", true},
		{"extra field", "ApiKey test-key extra", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			headers := http.Header{}
			if tc.header != "" {
				headers.Set("Authorization", tc.header)
			}
			got, err := GetAPIKey(headers)
			if (err != nil) != tc.wantError || got != tc.want {
				t.Fatalf("got %q, error %v; want %q, error %v", got, err, tc.want, tc.wantError)
			}
		})
	}
}
