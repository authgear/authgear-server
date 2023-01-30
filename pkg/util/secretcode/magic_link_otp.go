package secretcode

import (
	"crypto/subtle"
	"strings"

	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

var MagicLinkOTPSecretCode = MagicLinkOTPSecretCodeType{}

type MagicLinkOTPSecretCodeType struct{}

func (MagicLinkOTPSecretCodeType) Generate() string {
	code := rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)
	return code
}

func (MagicLinkOTPSecretCodeType) Compare(a, b string) bool {
	formattedCode := strings.TrimSpace(a)
	targetCode := strings.TrimSpace(b)
	return subtle.ConstantTimeCompare([]byte(formattedCode), []byte(targetCode)) == 1
}

func (MagicLinkOTPSecretCodeType) CheckFormat(value interface{}) error {
	return nil
}
