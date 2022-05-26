package secretcode

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"image"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type TOTP struct {
	Secret string
}

type QRCodeImageOptions struct {
	Issuer      string
	AccountName string
	Width       int
	Height      int
}

// https://github.com/google/google-authenticator/wiki/Key-Uri-Format#secret
// Base32 encoding as specified by RFC3548 (RFC4648) without padding.
var b32NoPadding = base32.StdEncoding.WithPadding(base32.NoPadding)

var validateOptsTOTP = totp.ValidateOpts{
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

// NewTOTPSecretFromRNG generates random TOTP secret encoded in Base32 without Padding.
func NewTOTPFromRNG() (*TOTP, error) {
	// https://tools.ietf.org/html/rfc4226#section-4
	// The RFC recommends a secret length of 160 bits.
	// That is 20 bytes.
	secretSize := 20
	secretBytes := make([]byte, secretSize)
	_, err := rand.Read(secretBytes)
	if err != nil {
		return &TOTP{Secret: ""}, err
	}
	secret := b32NoPadding.EncodeToString(secretBytes)
	return &TOTP{Secret: secret}, nil
}

func NewTOTPFromSecret(secret string) *TOTP {
	return &TOTP{Secret: secret}
}

// GenerateCode generates the TOTP code against the secret at the given time t.
func (c *TOTP) GenerateCode(t time.Time) (string, error) {
	return totp.GenerateCodeCustom(c.Secret, t, validateOptsTOTP)
}

// ValidateCode validates the TOTP code against the secret at the given time t.
func (c *TOTP) ValidateCode(t time.Time, code string) bool {
	formattedCode := strings.TrimSpace(code)
	ok, err := totp.ValidateCustom(formattedCode, c.Secret, t, validateOptsTOTP)
	if err != nil {
		return false
	}
	return ok
}

func (c *TOTP) QRCodeImage(opts QRCodeImageOptions) (image.Image, error) {
	q := url.Values{}
	q.Set("secret", c.Secret)
	q.Set("issuer", opts.Issuer)
	q.Set("algorithm", otp.AlgorithmSHA1.String())
	q.Set("digits", otp.DigitsSix.String())
	q.Set("period", strconv.FormatUint(uint64(validateOptsTOTP.Period), 10))
	u := &url.URL{
		Scheme:   "otpauth",
		Host:     "totp",
		Path:     fmt.Sprintf("/%s:%s", opts.Issuer, opts.AccountName),
		RawQuery: q.Encode(),
	}

	key, err := otp.NewKeyFromURL(u.String())
	if err != nil {
		return nil, err
	}

	return key.Image(opts.Width, opts.Height)
}
