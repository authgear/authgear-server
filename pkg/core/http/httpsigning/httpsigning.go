package httpsigning

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
)

const (
	// SigningAlgorithm is "HMAC-SHA256".
	SigningAlgorithm = "HMAC-SHA256"
)

const (
	// TimeFormat is YYYYMMDD'T'HHMMSS'Z'
	TimeFormat = "2006-01-02T15:04:05Z"
)

const (
	signedHeaders = "host"
)

var (
	ErrInvalidSignature = errors.New("invalid signature")
	ErrExpiredSignature = errors.New("expired signature")
)

func IsSigned(r *http.Request) bool {
	q := r.URL.Query()
	_, ok := q["x-skygear-signature"]
	return ok
}

// Sign turns r into a signed http.Request.
func Sign(key []byte, r *http.Request, t time.Time, expires int) {
	q := r.URL.Query()
	q.Set("x-skygear-algorithm", SigningAlgorithm)
	q.Set("x-skygear-date", FormatTime(t))
	q.Set("x-skygear-expires", strconv.Itoa(expires))
	q.Set("x-skygear-signedheaders", signedHeaders)
	r.URL.RawQuery = q.Encode()

	stringToSign := StringToSign(r, t)

	q = r.URL.Query()
	q.Set("x-skygear-signature", Hex(HMACSHA256(key, stringToSign)))
	r.URL.RawQuery = q.Encode()
}

func Verify(key []byte, r *http.Request, now time.Time) error {
	// Verify expires
	q := r.URL.Query()

	xSkygearSignature := q.Get("x-skygear-signature")
	claimedSig, err := hex.DecodeString(xSkygearSignature)
	if err != nil {
		return ErrInvalidSignature
	}

	xSkygearDate := q.Get("x-skygear-date")
	t, err := time.Parse(TimeFormat, xSkygearDate)
	if err != nil {
		return ErrInvalidSignature
	}

	xSkygearExpires := q.Get("x-skygear-expires")
	expires, err := strconv.Atoi(xSkygearExpires)
	if err != nil {
		return ErrInvalidSignature
	}

	stringToSign := StringToSign(r, t)
	mac := hmac.New(sha256.New, key)
	mac.Write(stringToSign)
	expectedSig := mac.Sum(nil)

	eq := hmac.Equal(claimedSig, expectedSig)
	if !eq {
		return ErrInvalidSignature
	}

	if now.After(t.Add(time.Duration(expires) * time.Second)) {
		return ErrExpiredSignature
	}

	return nil
}

// Hex is hex.EncodeToString.
func Hex(b []byte) string {
	return hex.EncodeToString(b)
}

// SHA256 is sha256.Sum256.
func SHA256(b []byte) []byte {
	arr := sha256.Sum256(b)
	return arr[:]
}

// HMACSHA256 is HMAC-SHA256.
func HMACSHA256(key []byte, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

// FormatTime formats t as TimeFormat.
func FormatTime(t time.Time) string {
	return t.Format(TimeFormat)
}

// StringToSign computes StringToSign.
func StringToSign(r *http.Request, t time.Time) []byte {
	s := fmt.Sprintf("%s\n%s\n%s", SigningAlgorithm, FormatTime(t), HashedCanonicalRequest(r))
	return []byte(s)
}

// HashedCanonicalRequest is HEX(SHA256(CanonicalRequest(r))).
func HashedCanonicalRequest(r *http.Request) string {
	return Hex(SHA256(CanonicalRequest(r)))
}

// CanonicalRequest turns r into bytes.
func CanonicalRequest(r *http.Request) []byte {
	buf := &bytes.Buffer{}

	buf.WriteString(r.Method)
	buf.WriteRune('\n')

	buf.WriteString(r.URL.EscapedPath())
	buf.WriteRune('\n')

	q := CanonicalQueryString(r.URL)
	buf.WriteString(q)
	buf.WriteRune('\n')

	canonicalHeaders := CanonicalHeaders(r)
	buf.WriteString(canonicalHeaders)
	buf.WriteRune('\n')

	buf.WriteRune('\n')

	buf.WriteString(signedHeaders)
	buf.WriteRune('\n')

	buf.WriteString("UNSIGNED-PAYLOAD")

	return buf.Bytes()
}

// CanonicalQueryString turns the query string of u into string.
func CanonicalQueryString(u *url.URL) string {
	q := u.Query()
	names := make([]string, len(q))

	i := 0
	for name := range q {
		names[i] = name
		i++
	}

	sort.Strings(names)

	buf := &strings.Builder{}
	for _, name := range names {
		if name == "" {
			continue
		}
		if name == "x-skygear-signature" {
			continue
		}
		values := q[name]
		for _, value := range values {
			if buf.Len() != 0 {
				buf.WriteRune('&')
			}
			buf.WriteString(url.QueryEscape(name))
			if value != "" {
				buf.WriteRune('=')
				buf.WriteString(url.QueryEscape(value))
			}
		}
	}

	return buf.String()
}

// CanonicalHeaders computes CANONICAL_HEADERS and SIGNED_HEADERS.
// Currently only Host is signed.
func CanonicalHeaders(r *http.Request) (canonicalHeaders string) {
	canonicalHeaders = fmt.Sprintf("host:%s", coreHttp.GetHost(r))
	return
}
