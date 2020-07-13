package otp

import (
	"crypto/rand"
	"encoding/base32"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// https://github.com/google/google-authenticator/wiki/Key-Uri-Format#secret
// Base32 encoding as specified by RFC3548 (RFC4648) without padding.
var b32NoPadding = base32.StdEncoding.WithPadding(base32.NoPadding)

var validateOpts = totp.ValidateOpts{
	// https://github.com/google/google-authenticator/wiki/Key-Uri-Format#period
	// The value must be 30
	Period: 30,
	// +- 1 period is good enough.
	Skew: 1,
	// https://github.com/google/google-authenticator/wiki/Key-Uri-Format#digits
	// The value must be 6
	Digits: otp.DigitsSix,
	// https://github.com/google/google-authenticator/wiki/Key-Uri-Format#algorithm
	// The value must be SHA1
	Algorithm: otp.AlgorithmSHA1,
}

// GenerateTOTPSecret generates random TOTP secret encoded in Base32 without Padding.
func GenerateTOTPSecret() (string, error) {
	// https://tools.ietf.org/html/rfc4226#section-4
	// The RFC recommends a secret length of 160 bits.
	// That is 20 bytes.
	secretSize := 20
	secretBytes := make([]byte, secretSize)
	_, err := rand.Read(secretBytes)
	if err != nil {
		return "", err
	}
	secret := b32NoPadding.EncodeToString(secretBytes)
	return secret, nil
}

// ValidateTOTP validates the TOTP code against the secret at the given time t.
func ValidateTOTP(secret string, code string, t time.Time) bool {
	ok, err := totp.ValidateCustom(code, secret, t, validateOpts)
	if err != nil {
		return false
	}
	return ok
}

// GenerateTOTP generates the TOTP code against the secret at the given time t.
func GenerateTOTP(secret string, t time.Time) (string, error) {
	return totp.GenerateCodeCustom(secret, t, validateOpts)
}
