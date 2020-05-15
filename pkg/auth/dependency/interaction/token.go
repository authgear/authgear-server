package interaction

import (
	"github.com/skygeario/skygear-server/pkg/core/base32"
	"github.com/skygeario/skygear-server/pkg/core/rand"
)

const (
	interactionTokenLength = 64
)

func generateToken() string {
	token := rand.StringWithAlphabet(interactionTokenLength, base32.Alphabet, rand.SecureRand)
	return token
}
