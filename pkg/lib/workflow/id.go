package workflow

import (
	"math/rand"

	"github.com/authgear/authgear-server/pkg/util/base32"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
)

const (
	idAlphabet string = base32.Alphabet
	idLength   int    = 32
)

var rng *rand.Rand = corerand.SecureRand

func newWorkflowID() string {
	return corerand.StringWithAlphabet(idLength, idAlphabet, rng)
}

func newInstanceID() string {
	return corerand.StringWithAlphabet(idLength, idAlphabet, rng)
}
