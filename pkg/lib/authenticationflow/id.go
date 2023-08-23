package authenticationflow

import (
	"fmt"
	"math/rand"

	"github.com/authgear/authgear-server/pkg/util/base32"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
)

const (
	idAlphabet string = base32.Alphabet
	idLength   int    = 32
)

var rng *rand.Rand = corerand.SecureRand

func newFlowID() string {
	return fmt.Sprintf("flowparent_%v", corerand.StringWithAlphabet(idLength, idAlphabet, rng))
}

func newInstanceID() string {
	return fmt.Sprintf("flow_%v", corerand.StringWithAlphabet(idLength, idAlphabet, rng))
}

func NewUserAgentID() string {
	return fmt.Sprintf("flowua_%v", corerand.StringWithAlphabet(idLength, idAlphabet, rng))
}
