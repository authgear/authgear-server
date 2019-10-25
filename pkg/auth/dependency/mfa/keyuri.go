package mfa

import (
	"bytes"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"net/url"
	"strconv"
	"strings"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
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

func NewKeyURI(issuer, accountName, secret string) *KeyURI {
	return &KeyURI{
		Type:        KeyURITypeTOTP,
		Issuer:      issuer,
		AccountName: accountName,
		Secret:      secret,
		Algorithm:   KeyURIAlgorithmSHA1,
		Digits:      6,
		Counter:     "",
		Period:      30,
	}
}

func (u *KeyURI) IsGoogleAuthenticatorCompatible() bool {
	return u.Type == KeyURITypeTOTP && u.Algorithm == KeyURIAlgorithmSHA1 && u.Digits == 6 && u.Period == 30
}

func (u *KeyURI) String() string {
	var path string
	if u.Issuer == "" {
		path = fmt.Sprintf("%s", url.PathEscape(u.AccountName))
	} else {
		path = fmt.Sprintf("%s:%s", url.PathEscape(u.Issuer), url.PathEscape(u.AccountName))
	}
	buf := &strings.Builder{}
	buf.WriteString("secret=")
	buf.WriteString(url.QueryEscape(u.Secret))
	if u.Issuer != "" {
		buf.WriteString("&issuer=")
		buf.WriteString(url.QueryEscape(u.Issuer))
	}
	return fmt.Sprintf("otpauth://totp/%s?%s", path, buf.String())
}

func (u *KeyURI) QRCodeDataURI() (string, error) {
	img, err := qr.Encode(u.String(), qr.M, qr.Auto)
	if err != nil {
		return "", err
	}

	img, err = barcode.Scale(img, 512, 512)
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	err = png.Encode(buf, img)
	if err != nil {
		return "", err
	}

	dataURIBuf := &bytes.Buffer{}
	dataURIBuf.WriteString("data:image/png;base64,")

	encoder := base64.NewEncoder(base64.StdEncoding, dataURIBuf)
	_, err = encoder.Write(buf.Bytes())
	if err != nil {
		return "", err
	}
	encoder.Close()

	return dataURIBuf.String(), nil
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
