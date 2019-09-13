package mfa

import (
	"encoding/base32"
	"errors"
	"net/url"
	"strconv"
	"strings"
)

var (
	ErrKeyURIInvalidScheme    = errors.New("keyuri: invalid scheme")
	ErrKeyURIHasPort          = errors.New("keyuri: no port is allowed")
	ErrKeyURIInvalidType      = errors.New("keyuri: invalid type")
	ErrKeyURIMismatchIssuer   = errors.New("keyuri: mismatch issuer")
	ErrKeyURIInvalidAlgorithm = errors.New("keyuri: invalid algorithm")
	ErrKeyURIInvalidDigits    = errors.New("keyuri: invalid digits")
	ErrKeyURIMissingCounter   = errors.New("keyuri: missing counter")
)

// KeyURIType is https://github.com/google/google-authenticator/wiki/Key-Uri-Format#types
type KeyURIType string

const (
	// KeyURITypeHOTP is "hotp"
	KeyURITypeHOTP = "hotp"
	// KeyURITypeTOTP is "totp"
	KeyURITypeTOTP = "totp"
)

// KeyURIAlgorithm is https://github.com/google/google-authenticator/wiki/Key-Uri-Format#algorithm
type KeyURIAlgorithm string

const (
	// KeyURIAlgorithmSHA1 is "SHA1"
	KeyURIAlgorithmSHA1 = "SHA1"
	// KeyURIAlgorithmSHA256 is "SHA256"
	KeyURIAlgorithmSHA256 = "SHA256"
	// KeyURIAlgorithmSHA512 is "SHA512"
	KeyURIAlgorithmSHA512 = "SHA512"
)

// KeyURI represents a key URI
// See https://github.com/google/google-authenticator/wiki/Key-Uri-Format
type KeyURI struct {
	Type        KeyURIType
	Issuer      string
	AccountName string
	Secret      string
	Algorithm   KeyURIAlgorithm
	Digits      int
	Counter     string
	Period      int
}

func (u *KeyURI) IsGoogleAuthenticatorCompatible() bool {
	return u.Type == KeyURITypeTOTP && u.Algorithm == KeyURIAlgorithmSHA1 && u.Digits == 6 && u.Period == 30
}

// ParseKeyURI parses s into KeyURI.
func ParseKeyURI(s string) (*KeyURI, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "otpauth" {
		return nil, ErrKeyURIInvalidScheme
	}

	port := u.Port()
	if port != "" {
		return nil, ErrKeyURIHasPort
	}

	typ := u.Hostname()
	if typ != KeyURITypeHOTP && typ != KeyURITypeTOTP {
		return nil, ErrKeyURIInvalidType
	}

	var pathIssuer string
	var accountName string
	label := strings.TrimPrefix(u.Path, "/")
	i := strings.IndexRune(label, ':')
	if i == -1 {
		accountName = label
	} else {
		pathIssuer = label[:i]
		accountName = label[i+1:]
	}

	values, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil, err
	}

	secret := values.Get("secret")
	_, err = base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return nil, err
	}

	var issuer string
	queryIssuer := values.Get("issuer")
	if queryIssuer != "" {
		if pathIssuer == "" {
			issuer = queryIssuer
		} else if queryIssuer != pathIssuer {
			return nil, ErrKeyURIMismatchIssuer
		} else {
			issuer = queryIssuer
		}
	} else {
		issuer = pathIssuer
	}

	algorithm := values.Get("algorithm")
	if algorithm == "" {
		algorithm = KeyURIAlgorithmSHA1
	} else if algorithm != KeyURIAlgorithmSHA1 && algorithm != KeyURIAlgorithmSHA256 && algorithm != KeyURIAlgorithmSHA512 {
		return nil, ErrKeyURIInvalidAlgorithm
	}

	var digitsValue int
	digits := values.Get("digits")
	if digits == "" {
		digits = "6"
	} else if digits != "6" && digits != "8" {
		return nil, ErrKeyURIInvalidDigits
	}
	digitsValue, _ = strconv.Atoi(digits)

	counter := values.Get("counter")
	if typ == KeyURITypeHOTP && counter == "" {
		return nil, ErrKeyURIMissingCounter
	}

	var periodValue int
	period := values.Get("period")
	if period == "" {
		periodValue = 30
	} else {
		periodValue, err = strconv.Atoi(period)
		if err != nil {
			return nil, err
		}
	}

	return &KeyURI{
		Type:        KeyURIType(typ),
		Issuer:      issuer,
		AccountName: accountName,
		Secret:      secret,
		Algorithm:   KeyURIAlgorithm(algorithm),
		Digits:      digitsValue,
		Counter:     counter,
		Period:      periodValue,
	}, nil
}
