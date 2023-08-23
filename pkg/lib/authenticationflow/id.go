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
	return fmt.Sprintf("authflowparent_%v", corerand.StringWithAlphabet(idLength, idAlphabet, rng))
}

func newInstanceID() string {
	return fmt.Sprintf("authflow_%v", corerand.StringWithAlphabet(idLength, idAlphabet, rng))
}

func NewUserAgentID() string {
	return fmt.Sprintf("authflowua_%v", corerand.StringWithAlphabet(idLength, idAlphabet, rng))
}
