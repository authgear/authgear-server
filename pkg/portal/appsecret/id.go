package appsecret

import (
	"math/rand"

	"github.com/authgear/authgear-server/pkg/util/base32"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
)

const (
	idAlphabet string = base32.Alphabet
	idLength   int    = 64
)

var rng *rand.Rand = corerand.SecureRand

func newSecretVisitTokenID() string {
	return corerand.StringWithAlphabet(idLength, idAlphabet, rng)
}
