package identity

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("a-strong-password")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "a-strong-password" {
		t.Fatal("HashPassword() returned plaintext")
	}
	if !VerifyPassword(hash, "a-strong-password") {
		t.Fatal("VerifyPassword() = false, want true")
	}
	if VerifyPassword(hash, "another-password") {
		t.Fatal("VerifyPassword() = true for another password")
	}
}

func TestHashPasswordUsesDifferentSalt(t *testing.T) {
	first, err := HashPassword("a-strong-password")
	if err != nil {
		t.Fatal(err)
	}
	second, err := HashPassword("a-strong-password")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("HashPassword() reused a salt")
	}
}

func TestHashPasswordRejectsShortPassword(t *testing.T) {
	if _, err := HashPassword("too-short"); err == nil {
		t.Fatal("HashPassword() error = nil, want weak password rejection")
	}
}

func TestVerifyPasswordRejectsMalformedHash(t *testing.T) {
	if VerifyPassword("not-an-argon2id-hash", "a-strong-password") {
		t.Fatal("VerifyPassword() accepted malformed hash")
	}
}
