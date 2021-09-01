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

func (t RecoveryCodeType) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	formatted, err := t.FormatForComparison(str)
	if err != nil {
		return fmt.Errorf("invalid recovery code: %w", err)
	}

	codeLength := len(formatted)
	if codeLength != 10 {
		return fmt.Errorf("unexpected recovery code length: %v", codeLength)
	}

	return nil
}
