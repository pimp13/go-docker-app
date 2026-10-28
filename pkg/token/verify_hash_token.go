package token

import (
	"crypto/sha256"
	"encoding/hex"
)

func VerifyHashToken(token, storedHash string) bool {
	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])

	return hashedToken == storedHash
}
