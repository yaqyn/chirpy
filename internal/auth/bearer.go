package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	if len(headers.Values("Authorization")) != 1 {
		return "", errors.New("exactly one authorization header is required")
	}
	parts := strings.Fields(headers.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("missing or invalid bearer token")
	}
	return parts[1], nil
}
