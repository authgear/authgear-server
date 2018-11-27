package userverify

import (
	"math/rand"
	"time"
)

type CodeGenerator interface {
	Generate() string
}

const (
	digits     = "0123456789"
	asciiLower = "abcdefghijklmnopqrstuvwxyz"
)

type defaultCodeGenerator struct {
	length  int
	charset string
}

func NewCodeGenerator(codeFormat string) CodeGenerator {
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
