package secretcode

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

var RecoveryCode = RecoveryCodeType{}

type RecoveryCodeType struct{}

func (RecoveryCodeType) Generate() string {
	code := rand.StringWithAlphabet(10, base32.Alphabet, rand.SecureRand)
	return code
}

func (RecoveryCodeType) FormatForHuman(code string) (formatted string) {
	halfLength := len(code) / 2
	formatted = fmt.Sprintf("%s-%s", code[:halfLength], code[halfLength:])
	return
}

func (RecoveryCodeType) FormatForComparison(code string) (formatted string, err error) {
	formatted, err = base32.Normalize(code)
	if err != nil {
		return
	}
	return
}
