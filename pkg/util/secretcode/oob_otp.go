package secretcode

import (
	"crypto/subtle"
	"fmt"

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

func (OOBOTPSecretCodeType) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	codeLength := len(str)
	if codeLength != 6 {
		return fmt.Errorf("unexpected OOB OTP code length: %v", codeLength)
	}

	for i, r := range str {
		if r < '0' || r > '9' {
			return fmt.Errorf("unexpected OOB OTP code character at index %v: %#v", i, string(r))
		}
	}

	return nil
}
