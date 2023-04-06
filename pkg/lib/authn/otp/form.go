package otp

import "github.com/authgear/authgear-server/pkg/util/secretcode"

type Form string

const (
	FormCode Form = "code"
	FormLink Form = "link"
)

func (f Form) AllowLookupByCode() bool {
	return f == FormLink
}

func (f Form) codeType() secretCode {
	switch f {
	case FormCode:
		return secretcode.OOBOTPSecretCode
	case FormLink:
		return secretcode.LinkOTPSecretCode
	default:
		panic("unexpected form: " + f)
	}
}

func (f Form) GenerateCode() string {
	return f.codeType().Generate()
}

func (f Form) VerifyCode(input string, expected string) bool {
	return f.codeType().Compare(input, expected)
}

type secretCode interface {
	Generate() string
	Compare(string, string) bool
}
