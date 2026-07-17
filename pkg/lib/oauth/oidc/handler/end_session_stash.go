package handler

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc/protocol"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// endSessionRefQueryParam carries the sealed, opaque request blob across the
// POST -> redirect-to-self -> GET round trip. It is safe to appear in the URL
// because it contains no plaintext PII; only the holder of the matching
// EndSessionStashCookieDef cookie (set on the same origin, same response) can
// decrypt it.
const endSessionRefQueryParam = "x_end_session_ref"

const endSessionStashCookieMaxAge = 300 // 5 minutes; only needs to survive one redirect round trip.

var EndSessionStashCookieDef = &httputil.CookieDef{
	NameSuffix: "end_session_stash",
	Path:       "/oauth2/end_session",
	SameSite:   http.SameSiteLaxMode,
	MaxAge:     endSessionStashCookieMaxAgePtr(),
}

func endSessionStashCookieMaxAgePtr() *int {
	v := endSessionStashCookieMaxAge
	return &v
}

// ErrEndSessionStashInvalid is returned when a resumed request carries a
// x_end_session_ref query value that cannot be opened: the cookie is missing
// (expired, blocked, or a different browser/tab), the key/ciphertext pairing
// doesn't authenticate (tampered or mismatched), or the payload doesn't
// decode. EndSessionHandler.resumeFromStash (handler_end_session.go) logs
// this and falls back to treating the request as if it had no parameters at
// all, rather than surfacing it as an error: reaching this state in normal
// use (e.g. revisiting a stale link from browser history after the
// short-lived stash cookie has expired) is expected, not exceptional.
var ErrEndSessionStashInvalid = errors.New("end_session: invalid or expired stash")

// sealEndSessionRequest encrypts req under a freshly generated random 256-bit
// key using AES-GCM. It returns the key (to be stored in
// EndSessionStashCookieDef) and the sealed blob nonce||ciphertext||tag, both
// base64url-encoded (to be carried in endSessionRefQueryParam).
func sealEndSessionRequest(req protocol.EndSessionRequest) (key string, sealed string, err error) {
	keyBytes := make([]byte, 32)
	if _, err = rand.Read(keyBytes); err != nil {
		return "", "", err
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", "", err
	}

	plaintext, err := json.Marshal(req)
	if err != nil {
		return "", "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	key = base64.RawURLEncoding.EncodeToString(keyBytes)
	sealed = base64.RawURLEncoding.EncodeToString(ciphertext)
	return key, sealed, nil
}

// openEndSessionRequest reverses sealEndSessionRequest. Any failure (bad
// base64, bad key length, GCM authentication failure, bad JSON) collapses to
// ErrEndSessionStashInvalid; callers must not distinguish further, since the
// distinction is not actionable by the caller.
func openEndSessionRequest(key string, sealed string) (protocol.EndSessionRequest, error) {
	keyBytes, err := base64.RawURLEncoding.DecodeString(key)
	if err != nil {
		return nil, ErrEndSessionStashInvalid
	}
	ciphertext, err := base64.RawURLEncoding.DecodeString(sealed)
	if err != nil {
		return nil, ErrEndSessionStashInvalid
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, ErrEndSessionStashInvalid
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrEndSessionStashInvalid
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrEndSessionStashInvalid
	}
	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, ErrEndSessionStashInvalid
	}

	var req protocol.EndSessionRequest
	if err := json.Unmarshal(plaintext, &req); err != nil {
		return nil, ErrEndSessionStashInvalid
	}
	return req, nil
}
