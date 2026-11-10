package sso

import (
	"crypto/sha256"
	"encoding/base64"
	"test/internal/model"
	"testing"
)

func TestVerifyPKCES256(t *testing.T) {
	verifier := "demo-verifier-value"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])

	code := model.SSOAuthorizationCode{
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
	}
	if err := verifyPKCE(code, verifier); err != nil {
		t.Fatal(err)
	}
	if err := verifyPKCE(code, "wrong"); err == nil {
		t.Fatal("expected pkce failure")
	}
}

func TestNormalizeScopes(t *testing.T) {
	if got := normalizeScopes(""); got != defaultScope {
		t.Fatalf("got = %q", got)
	}
	if got := normalizeScopes(" email openid email "); got != "email openid" {
		t.Fatalf("got = %q", got)
	}
}
