package auth

import (
	"encoding/hex"
	"testing"
)

func TestMakeRefreshToken(t *testing.T) {
	first := MakeRefreshToken()
	data, err := hex.DecodeString(first)
	if err != nil || len(data) != 32 {
		t.Fatalf("expected 32 hex-encoded bytes; length %d, error %v", len(data), err)
	}
	if first == MakeRefreshToken() {
		t.Fatal("refresh tokens should be different")
	}
}
