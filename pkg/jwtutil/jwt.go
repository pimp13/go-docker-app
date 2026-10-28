package jwtutil

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalid = errors.New("invalid token")
	ErrExpired = errors.New("token expired")
)

// Claims is the signed JWT payload for an authenticated user.
type Claims struct {
	Role                string `json:"role"`
	PasswordFingerprint string `json:"pfp"`
	jwt.RegisteredClaims
}

// IssueParams controls how an access token is signed.
type IssueParams struct {
	Secret              string
	Issuer              string
	Subject             string
	Role                string
	PasswordFingerprint string
	TTL                 time.Duration
	Now                 time.Time
}

// Issue signs an HS256 JWT and returns the compact token plus its expiry.
func Issue(p IssueParams) (string, time.Time, error) {
	if p.Secret == "" || p.Subject == "" || p.PasswordFingerprint == "" || p.TTL <= 0 {
		return "", time.Time{}, ErrInvalid
	}

	now := p.Now.UTC()
	if p.Now.IsZero() {
		now = time.Now().UTC()
	}

	expiresAt := now.Add(p.TTL)
	claims := Claims{
		Role:                p.Role,
		PasswordFingerprint: p.PasswordFingerprint,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    p.Issuer,
			Subject:   p.Subject,
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
	if claims.ExpiresAt == nil {
		return "", time.Time{}, ErrInvalid
	}
	return signed, claims.ExpiresAt.Time, nil
}

// Parse verifies the signature, issuer, and expiry of an access token.
func Parse(secret, issuer, raw string) (*Claims, error) {
	if strings.TrimSpace(raw) == "" || secret == "" {
		return nil, ErrInvalid
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if t.Method == nil || t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, ErrInvalid
		}
		return []byte(secret), nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(5*time.Second),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpired
		}
		return nil, ErrInvalid
	}
	if token == nil || !token.Valid || claims.Subject == "" || claims.PasswordFingerprint == "" {
		return nil, ErrInvalid
	}
	return claims, nil
}
