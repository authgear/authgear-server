package otp

import (
	"crypto/subtle"

	"github.com/authgear/authgear-server/pkg/core/base32"
	"github.com/authgear/authgear-server/pkg/core/rand"
)

type Format struct {
	Alphabet string
	Length   int
}

var (
	FormatNumeric = &Format{
		Alphabet: "0123456789",
		Length:   4,
	}
	FormatComplex = &Format{
		Alphabet: base32.Alphabet,
		Length:   8,
	}
)

func (f *Format) Generate() string {
	code := rand.StringWithAlphabet(f.Length, f.Alphabet, rand.SecureRand)
	return code
}

func ValidateOTP(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
