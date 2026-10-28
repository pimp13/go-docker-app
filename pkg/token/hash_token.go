package token

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashToken(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(tokenHash[:])
}
