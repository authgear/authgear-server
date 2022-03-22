package httpsigning

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/httputil"
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

var InvalidSignatureQueryParam = apierrors.Invalid.WithReason("InvalidSignatureQueryParam")
var InvalidSignature = apierrors.Invalid.WithReason("InvalidSignature")
var ErrExpiredSignature = apierrors.Invalid.WithReason("ExpiredSignature").New("expired signature")

// Sign turns r into a signed http.Request.
func Sign(key []byte, r *http.Request, t time.Time, expires int) {
	q := r.URL.Query()
	q.Set("x-authgear-algorithm", SigningAlgorithm)
	q.Set("x-authgear-date", FormatTime(t))
	q.Set("x-authgear-expires", strconv.Itoa(expires))
	q.Set("x-authgear-signedheaders", signedHeaders)
	r.URL.RawQuery = q.Encode()

	host := httputil.HTTPHost(r.URL.Host)
	stringToSign, _ := StringToSign(host, r, t)

	q = r.URL.Query()
	q.Set("x-authgear-signature", Hex(HMACSHA256(key, stringToSign)))
	r.URL.RawQuery = q.Encode()
}

func Verify(key []byte, host httputil.HTTPHost, r *http.Request, now time.Time) error {
	// Verify expires
	q := r.URL.Query()

	xSignature := q.Get("x-authgear-signature")
	claimedSig, err := hex.DecodeString(xSignature)
	if err != nil {
		return InvalidSignatureQueryParam.New("invalid x-authgear-signature")
	}

	xDate := q.Get("x-authgear-date")
	t, err := time.Parse(TimeFormat, xDate)
	if err != nil {
		return InvalidSignatureQueryParam.New("invalid x-authgear-date")
	}

	xExpires := q.Get("x-authgear-expires")
	expires, err := strconv.Atoi(xExpires)
	if err != nil {
		return InvalidSignatureQueryParam.New("invalid x-authgear-expires")
	}

	stringToSign, canonicalRequest := StringToSign(host, r, t)
	mac := hmac.New(sha256.New, key)
	mac.Write(stringToSign)
	expectedSig := mac.Sum(nil)

	eq := hmac.Equal(claimedSig, expectedSig)
	if !eq {
		return InvalidSignature.NewWithInfo("invalid signature", apierrors.Details{
			"canonical_request": string(canonicalRequest),
		})
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
func StringToSign(host httputil.HTTPHost, r *http.Request, t time.Time) (stringToSign []byte, canonicalRequest []byte) {
	hashded, canonicalRequest := HashedCanonicalRequest(host, r)
	s := fmt.Sprintf("%s\n%s\n%s", SigningAlgorithm, FormatTime(t), hashded)
	return []byte(s), canonicalRequest
}

// HashedCanonicalRequest is HEX(SHA256(CanonicalRequest(host, r))).
func HashedCanonicalRequest(host httputil.HTTPHost, r *http.Request) (hashed string, canonicalRequest []byte) {
	canonicalRequest = CanonicalRequest(host, r)
	hashed = Hex(SHA256(canonicalRequest))
	return
}

// CanonicalRequest turns r into bytes.
func CanonicalRequest(host httputil.HTTPHost, r *http.Request) []byte {
	buf := &bytes.Buffer{}

	buf.WriteString(r.Method)
	buf.WriteRune('\n')

	buf.WriteString(r.URL.EscapedPath())
	buf.WriteRune('\n')

	q := CanonicalQueryString(r.URL)
	buf.WriteString(q)
	buf.WriteRune('\n')

	canonicalHeaders := CanonicalHeaders(host)
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
		if name == "x-authgear-signature" {
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
func CanonicalHeaders(host httputil.HTTPHost) (canonicalHeaders string) {
	canonicalHeaders = fmt.Sprintf("host:%s", string(host))
	return
}
