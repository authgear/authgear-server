package userverify

import (
	"math/rand"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

const (
	digits     = "0123456789"
	asciiLower = "abcdefghijklmnopqrstuvwxyz"
)

type CodeGeneratorFactory interface {
	NewCodeGenerator(key string) CodeGenerator
}

type CodeGenerator interface {
	Generate() string
}

type DefaultCodeGeneratorFactory struct {
	CodeFormatMap map[string]config.UserVerificationCodeFormat
}

func NewDefaultCodeGeneratorFactory(c config.TenantConfiguration) CodeGeneratorFactory {
	userVerifyConfig := c.UserConfig.UserVerification
	f := DefaultCodeGeneratorFactory{
		CodeFormatMap: map[string]config.UserVerificationCodeFormat{},
	}
	for key, config := range userVerifyConfig.LoginIDKeys {
		f.CodeFormatMap[key] = config.CodeFormat
	}

	return &f
}

func (d *DefaultCodeGeneratorFactory) NewCodeGenerator(key string) CodeGenerator {
	return NewCodeGenerator(d.CodeFormatMap[key])
}

type defaultCodeGenerator struct {
	length  int
	charset string
}

func NewCodeGenerator(codeFormat config.UserVerificationCodeFormat) CodeGenerator {
	switch codeFormat {
	case "numeric":
		return &defaultCodeGenerator{
			length:  6,
			charset: digits,
		}
	case "complex":
		return &defaultCodeGenerator{
			length:  8,
			charset: digits + asciiLower,
		}
	}

	return nil
}

func (d *defaultCodeGenerator) Generate() string {
	return randomStringWithCharset(d.length, d.charset)
}

func randomStringWithCharset(length int, charset string) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}
