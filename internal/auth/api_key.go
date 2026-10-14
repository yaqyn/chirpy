package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	if len(headers.Values("Authorization")) != 1 {
		return "", errors.New("exactly one authorization header is required")
	}
	parts := strings.Fields(headers.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "ApiKey") {
		return "", errors.New("missing or invalid API key")
	}
	return parts[1], nil
}
