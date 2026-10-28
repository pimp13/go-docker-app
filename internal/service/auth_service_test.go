package service

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestDummyPasswordHashIsBcrypt(t *testing.T) {
	err := bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte("definitely-wrong"))
	if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Fatalf("dummy hash is not a valid bcrypt hash: %v", err)
	}
}

func TestPasswordFingerprintChangesWithHash(t *testing.T) {
	a := passwordFingerprint("hash-a")
	b := passwordFingerprint("hash-b")
	if a == "" || a == b {
		t.Fatalf("fingerprints = %q %q", a, b)
	}
	if !fingerprintMatches("hash-a", a) || fingerprintMatches("hash-a", b) {
		t.Fatal("fingerprint comparison failed")
	}
}
