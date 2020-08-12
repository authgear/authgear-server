package webapp

import (
	"github.com/authgear/authgear-server/pkg/util/base32"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
)

const (
	idAlphabet string = base32.Alphabet
	idLength   int    = 32
)

func NewID() string {
	return corerand.StringWithAlphabet(idLength, idAlphabet, corerand.SecureRand)
}
