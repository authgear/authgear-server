package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// HMACSHA256String returns the hex-encoded string of HMAC-SHA256 code of body using secret as key.
func HMACSHA256String(secret []byte, body []byte) string {
	hasher := hmac.New(sha256.New, secret)
	hasher.Write(body)
	signature := hasher.Sum(nil)
	return hex.EncodeToString(signature)
}
