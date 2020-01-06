package userverify

import (
	"github.com/skygeario/skygear-server/pkg/core/base32"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/rand"
)

type codeFormat struct {
	Length   int
	Alphabet string
}

var codeFormats = map[config.UserVerificationCodeFormat]codeFormat{
	config.UserVerificationCodeFormatNumeric: codeFormat{
		Length:   6,
		Alphabet: "0123456789",
	},
	config.UserVerificationCodeFormatComplex: codeFormat{
		Length:   8,
		Alphabet: base32.Alphabet,
	},
}

type CodeGenerator interface {
	Generate(loginIDKey string) string
}

type defaultCodeGenerator struct {
	LoginIDKeyCodeFormats map[string]config.UserVerificationCodeFormat
}

func NewCodeGenerator(c config.TenantConfiguration) CodeGenerator {
	formats := map[string]config.UserVerificationCodeFormat{}
	for _, config := range c.AppConfig.UserVerification.LoginIDKeys {
		formats[config.Key] = config.CodeFormat
	}

	return &defaultCodeGenerator{LoginIDKeyCodeFormats: formats}
}

func (d *defaultCodeGenerator) Generate(loginIDKey string) string {
	codeFormat := codeFormats[d.LoginIDKeyCodeFormats[loginIDKey]]
	return rand.StringWithAlphabet(codeFormat.Length, codeFormat.Alphabet, rand.SecureRand)
}
