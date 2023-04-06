package secretcode

import (
	"crypto/subtle"
	"strings"

	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

const linkOTPLength = 32

var LinkOTPSecretCode = LinkOTPSecretCodeType{}

type LinkOTPSecretCodeType struct{}

func (LinkOTPSecretCodeType) Length() int {
	return linkOTPLength
}

func (LinkOTPSecretCodeType) Generate() string {
	code := rand.StringWithAlphabet(linkOTPLength, base32.Alphabet, rand.SecureRand)
	return code
}

func (LinkOTPSecretCodeType) Compare(a, b string) bool {
	formattedCode := strings.TrimSpace(a)
	targetCode := strings.TrimSpace(b)
	return subtle.ConstantTimeCompare([]byte(formattedCode), []byte(targetCode)) == 1
}

func (LinkOTPSecretCodeType) CheckFormat(value interface{}) error {
	return nil
}
