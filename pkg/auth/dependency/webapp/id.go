package webapp

import (
	"github.com/authgear/authgear-server/pkg/core/base32"
	corerand "github.com/authgear/authgear-server/pkg/core/rand"
)

const (
	idAlphabet string = base32.Alphabet
	idLength   int    = 32
)

func NewID() string {
	return corerand.StringWithAlphabet(idLength, idAlphabet, corerand.SecureRand)
}
