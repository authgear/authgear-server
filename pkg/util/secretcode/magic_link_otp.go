package secretcode

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"strings"
)

var MagicLinkOTPSecretCode = MagicLinkOTPSecretCodeType{}

type MagicLinkOTPSecretCodeType struct{}

func (MagicLinkOTPSecretCodeType) Generate() string {
	code := make([]byte, 32)
	rand.Read(code)
	return hex.EncodeToString(code)
}

func (MagicLinkOTPSecretCodeType) Compare(a, b string) bool {
	formattedCode := strings.TrimSpace(a)
	targetCode := strings.TrimSpace(b)
	return subtle.ConstantTimeCompare([]byte(formattedCode), []byte(targetCode)) == 1
}

func (MagicLinkOTPSecretCodeType) CheckFormat(value interface{}) error {
	return nil
}
