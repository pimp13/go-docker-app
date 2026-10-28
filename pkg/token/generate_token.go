package token

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

// 24 hours
const TokenExpirationTime = 24 * time.Hour

type TokenPayload struct {
	Random    string `json:"random"`
	ExpiresAt int64  `json:"exp"`
}

// CreateToken creates a random, signed token.
func CreateToken(secret string) (string, error) {
	// Generate 32 cryptographically secure random bytes.
	randomBytes := make([]byte, 32)

	if _, err := io.ReadFull(rand.Reader, randomBytes); err != nil {
		return "", err
	}

	payload := TokenPayload{
		Random:    base64.RawURLEncoding.EncodeToString(randomBytes),
		ExpiresAt: time.Now().Add(TokenExpirationTime).Unix(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	// Sign the payload using HMAC-SHA256.
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encodedPayload))

	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return encodedPayload + "." + signature, nil
}

// ValidateToken verifies the signature and expiration.
func ValidateToken(token string, secret string) error {
	parts := strings.Split(token, ".")

	if len(parts) != 2 {
		return errors.New("invalid token format")
	}

	encodedPayload := parts[0]
	encodedSignature := parts[1]

	// Decode the provided signature.
	signature, err := base64.RawURLEncoding.DecodeString(encodedSignature)
	if err != nil {
		return errors.New("invalid signature encoding")
	}

	// Calculate the expected signature.
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encodedPayload))

	expectedSignature := mac.Sum(nil)

	// Compare signatures in constant time.
	if !hmac.Equal(signature, expectedSignature) {
		return errors.New("invalid token signature")
	}

	// Decode the payload only after verifying its signature.
	payloadBytes, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return errors.New("invalid payload encoding")
	}

	var payload TokenPayload

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return errors.New("invalid token payload")
	}

	if payload.ExpiresAt <= time.Now().Unix() {
		return errors.New("token expired")
	}

	return nil
}
