package jwtutil

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const TokenUseAccess = "access"

// SSOClaims is the access token issued to SSO clients after code exchange.
type SSOClaims struct {
	TokenUse    string   `json:"token_use"`
	Email       string   `json:"email,omitempty"`
	Name        string   `json:"name,omitempty"`
	Role        string   `json:"role,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

// SSOIssueParams controls SSO access-token issuance.
type SSOIssueParams struct {
	Secret      string
	Issuer      string
	Audience    string
	Subject     string
	Email       string
	Name        string
	Role        string
	Permissions []string
	TTL         time.Duration
	Now         time.Time
}

// IssueSSO signs an SSO access token bound to a client audience.
func IssueSSO(p SSOIssueParams) (string, time.Time, error) {
	if p.Secret == "" || p.Subject == "" || p.Audience == "" || p.TTL <= 0 {
		return "", time.Time{}, ErrInvalid
	}

	now := p.Now.UTC()
	if p.Now.IsZero() {
		now = time.Now().UTC()
	}
	expiresAt := now.Add(p.TTL)

	claims := SSOClaims{
		TokenUse:    TokenUseAccess,
		Email:       p.Email,
		Name:        p.Name,
		Role:        p.Role,
		Permissions: p.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    p.Issuer,
			Subject:   p.Subject,
			Audience:  []string{p.Audience},
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(p.Secret))
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

// ParseSSO verifies an SSO access token, optionally enforcing audience.
func ParseSSO(secret, issuer, audience, raw string) (*SSOClaims, error) {
	if strings.TrimSpace(raw) == "" || secret == "" {
		return nil, ErrInvalid
	}

	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(5 * time.Second),
	}
	if audience != "" {
		opts = append(opts, jwt.WithAudience(audience))
	}

	claims := &SSOClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if t.Method == nil || t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, ErrInvalid
		}
		return []byte(secret), nil
	}, opts...)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpired
		}
		return nil, ErrInvalid
	}
	if token == nil || !token.Valid || claims.Subject == "" || claims.TokenUse != TokenUseAccess {
		return nil, ErrInvalid
	}
	return claims, nil
}
