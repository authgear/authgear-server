package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// HMACSHA256 returns the HMAC-SHA256 code of body using secret as key.
func HMACSHA256(body []byte, secret []byte) string {
	hasher := hmac.New(sha256.New, secret)
	hasher.Write(body)
	signature := hasher.Sum(nil)
	return hex.EncodeToString(signature)
}
