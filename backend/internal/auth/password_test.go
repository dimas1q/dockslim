package auth

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	password := "strongpassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected hashing to succeed, got %v", err)
	}

	if err := VerifyPassword(hash, password); err != nil {
		t.Fatalf("expected verification to succeed, got %v", err)
	}

	if err := VerifyPassword(hash, "wrongpassword"); err == nil {
		t.Fatalf("expected verification to fail for wrong password")
	}
}
