package settingsaction

import (
	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

func NewSettingsActionID() string {
	return rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)
}
