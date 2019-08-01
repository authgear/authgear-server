package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256String returns the hex hash of s.
func SHA256String(s string) string {
	b := []byte(s)
	outArr := sha256.Sum256(b)
	outSlice := outArr[:]
	return hex.EncodeToString(outSlice)
}
