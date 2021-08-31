package secretcode

import (
	"crypto/subtle"

	"github.com/authgear/authgear-server/pkg/util/rand"
)

var OOBOTPSecretCode = OOBOTPSecretCodeType{}

type OOBOTPSecretCodeType struct{}

func (OOBOTPSecretCodeType) Generate() string {
	code := rand.StringWithAlphabet(6, "0123456789", rand.SecureRand)
	return code
}

func (OOBOTPSecretCodeType) Compare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
