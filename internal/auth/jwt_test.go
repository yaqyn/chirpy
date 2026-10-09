package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestJWT(t *testing.T) {
	userID := uuid.New()
	for _, tc := range []struct {
		name             string
		expiresIn        time.Duration
		validationSecret string
		wantError        bool
	}{
		{"valid", time.Hour, "test-secret", false},
		{"expired", -time.Hour, "test-secret", true},
		{"wrong secret", time.Hour, "wrong-secret", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			token, err := MakeJWT(userID, "test-secret", tc.expiresIn)
			if err != nil {
				t.Fatal(err)
			}
			got, err := ValidateJWT(token, tc.validationSecret)
			if tc.wantError {
				if err == nil {
					t.Fatal("expected invalid token to be rejected")
				}
				return
			}
			if err != nil || got != userID {
				t.Fatalf("got user %v, error %v; want %v", got, err, userID)
			}
		})
	}
}

func TestJWTRejectsInvalidClaimsAndMethods(t *testing.T) {
	for _, tc := range []struct {
		name    string
		method  jwt.SigningMethod
		issuer  string
		subject string
	}{
		{"wrong signing method", jwt.SigningMethodHS384, "chirpy-access", uuid.NewString()},
		{"wrong issuer", jwt.SigningMethodHS256, "other-app", uuid.NewString()},
		{"invalid user ID", jwt.SigningMethodHS256, "chirpy-access", "not-a-uuid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			token, err := jwt.NewWithClaims(tc.method, jwt.RegisteredClaims{
				Issuer: tc.issuer, Subject: tc.subject, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			}).SignedString([]byte("test-secret"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ValidateJWT(token, "test-secret"); err == nil {
				t.Fatal("expected token to be rejected")
			}
		})
	}
}
