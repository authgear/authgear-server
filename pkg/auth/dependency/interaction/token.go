package interaction

import (
	"github.com/authgear/authgear-server/pkg/core/base32"
	"github.com/authgear/authgear-server/pkg/core/rand"
)

const (
	interactionTokenLength = 64
)

func generateToken() string {
	token := rand.StringWithAlphabet(interactionTokenLength, base32.Alphabet, rand.SecureRand)
	return token
}
