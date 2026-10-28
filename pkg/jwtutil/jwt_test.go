package jwtutil

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestIssueAndParse(t *testing.T) {
	secret := strings.Repeat("k", 32)
	signed, exp, err := Issue(IssueParams{
		Secret:              secret,
		Issuer:              "test",
		Subject:             "user-id",
		Role:                "user",
		PasswordFingerprint: "abc",
		TTL:                 time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}

	claims, err := Parse(secret, "test", signed)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "user-id" || claims.Role != "user" || claims.PasswordFingerprint != "abc" {
		t.Fatalf("claims = %+v", claims)
	}
	if time.Until(exp) < 59*time.Minute {
		t.Fatalf("expiry too soon: %s", exp)
	}
}

func TestParseRejectsTamperedToken(t *testing.T) {
	secret := strings.Repeat("k", 32)
	signed, _, err := Issue(IssueParams{
		Secret:              secret,
		Issuer:              "test",
		Subject:             "user-id",
		Role:                "user",
		PasswordFingerprint: "abc",
		TTL:                 time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}

	parts := strings.Split(signed, ".")
	parts[2] = strings.Repeat("a", len(parts[2]))
	_, err = Parse(secret, "test", strings.Join(parts, "."))
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err = %v", err)
	}
}

func TestParseRejectsExpiredToken(t *testing.T) {
	secret := strings.Repeat("k", 32)
	signed, _, err := Issue(IssueParams{
		Secret:              secret,
		Issuer:              "test",
		Subject:             "user-id",
		Role:                "user",
		PasswordFingerprint: "abc",
		TTL:                 time.Minute,
		Now:                 time.Now().Add(-time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = Parse(secret, "test", signed)
	if !errors.Is(err, ErrExpired) {
		t.Fatalf("err = %v", err)
	}
}

func TestParseRejectsWrongIssuer(t *testing.T) {
	secret := strings.Repeat("k", 32)
	signed, _, err := Issue(IssueParams{
		Secret:              secret,
		Issuer:              "test",
		Subject:             "user-id",
		Role:                "user",
		PasswordFingerprint: "abc",
		TTL:                 time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = Parse(secret, "other", signed)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err = %v", err)
	}
}
