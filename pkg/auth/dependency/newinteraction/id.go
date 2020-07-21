package newinteraction

import (
	"github.com/authgear/authgear-server/pkg/core/base32"
	corerand "github.com/authgear/authgear-server/pkg/core/rand"
)

const (
	idAlphabet string = base32.Alphabet
	idLength   int    = 32
)

func newGraphID() string {
	return corerand.StringWithAlphabet(idLength, idAlphabet, corerand.SecureRand)
}

func newInstanceID() string {
	return corerand.StringWithAlphabet(idLength, idAlphabet, corerand.SecureRand)
}
