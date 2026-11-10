package jwtutil

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestIssueAndParseSSO(t *testing.T) {
	secret := strings.Repeat("s", 32)
	token, _, err := IssueSSO(SSOIssueParams{
		Secret:      secret,
		Issuer:      "test",
		Audience:    "demo-app",
		Subject:     "user-1",
		Email:       "ada@example.com",
		Name:        "Ada Lovelace",
		Role:        "user",
		Permissions: []string{"articles:read"},
		TTL:         time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}

	claims, err := ParseSSO(secret, "test", "demo-app", token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "user-1" || claims.Email != "ada@example.com" || claims.TokenUse != TokenUseAccess {
		t.Fatalf("claims = %+v", claims)
	}
}

func TestParseSSORejectsWrongAudience(t *testing.T) {
	secret := strings.Repeat("s", 32)
	token, _, err := IssueSSO(SSOIssueParams{
		Secret:   secret,
		Issuer:   "test",
		Audience: "demo-app",
		Subject:  "user-1",
		TTL:      time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = ParseSSO(secret, "test", "other-app", token)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err = %v", err)
	}
}
