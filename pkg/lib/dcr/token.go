package dcr

import (
	"github.com/authgear/authgear-server/pkg/util/crypto"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

const (
	tokenAlphabet     = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	tokenRandomLength = 22 // matches spec: "22 chars URL-safe base64 (16 bytes)"

	IATPrefixThirdParty = "iat_tp_"
	IATPrefixFirstParty = "iat_fp_"
)

// GenerateInitialAccessToken returns the plaintext token (returned to the
// caller once) and its SHA-256 hash (persisted).
func GenerateInitialAccessToken(t InitialAccessTokenType) (plaintext string, hash string) {
	prefix := IATPrefixThirdParty
	if t == InitialAccessTokenTypeFirstParty {
		prefix = IATPrefixFirstParty
	}
	plaintext = prefix + rand.StringWithAlphabet(tokenRandomLength, tokenAlphabet, rand.SecureRand)
	hash = crypto.SHA256String(plaintext)
	return
}

func HashInitialAccessToken(plaintext string) string {
	return crypto.SHA256String(plaintext)
}
