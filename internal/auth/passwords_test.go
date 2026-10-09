package auth

import "testing"

func TestPasswordHash(t *testing.T) {
	hash, err := HashPassword("test password")
	if err != nil {
		t.Fatal(err)
	}
	second, err := HashPassword("test password")
	if err != nil {
		t.Fatal(err)
	}
	if hash == second {
		t.Fatal("identical passwords should receive different salted hashes")
	}
	for _, tc := range []struct {
		password string
		want     bool
	}{
		{"test password", true}, {"wrong password", false},
	} {
		match, err := CheckPasswordHash(tc.password, hash)
		if err != nil || match != tc.want {
			t.Fatalf("password match = %v, error = %v; want %v", match, err, tc.want)
		}
	}
	if _, err := CheckPasswordHash("test password", "unset"); err == nil {
		t.Fatal("invalid stored hash should return an error")
	}
}
