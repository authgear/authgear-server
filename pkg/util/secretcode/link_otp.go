package secretcode

import (
	"context"
	// nolint:gosec
	"crypto/md5"
	"crypto/subtle"
	"fmt"
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

func (LinkOTPSecretCodeType) GenerateDeterministic(data string) string {
	b := []byte(data)
	hash := md5.Sum(b) // nolint:gosec
	return fmt.Sprintf("%x", hash)
}

func (LinkOTPSecretCodeType) Compare(a, b string) bool {
	formattedCode := strings.TrimSpace(a)
	targetCode := strings.TrimSpace(b)
	return subtle.ConstantTimeCompare([]byte(formattedCode), []byte(targetCode)) == 1
}

func (LinkOTPSecretCodeType) CheckFormat(ctx context.Context, value interface{}) error {
	return nil
}
