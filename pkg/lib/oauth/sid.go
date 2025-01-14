package oauth

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/authgear/authgear-server/pkg/lib/session"
)

type SessionLike interface {
	SessionID() string
	SessionType() session.Type
}

func EncodeSID(s SessionLike) string {
	return EncodeSIDByRawValues(s.SessionType(), s.SessionID())
}

func EncodeSIDByRawValues(sessionType session.Type, sessionID string) string {
	raw := fmt.Sprintf("%s:%s", sessionType, sessionID)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func DecodeSID(sid string) (typ session.Type, sessionID string, ok bool) {
	bytes, err := base64.RawURLEncoding.DecodeString(sid)
	if err != nil {
		return
	}

	if !utf8.Valid(bytes) {
		return
	}
	str := string(bytes)

	parts := strings.Split(str, ":")
	if len(parts) != 2 {
		return
	}

	typStr := parts[0]
	sessionID = parts[1]
	switch typStr {
	case string(session.TypeIdentityProvider):
		typ = session.TypeIdentityProvider
	case string(session.TypeOfflineGrant):
		typ = session.TypeOfflineGrant
	}
	if typ == "" {
		return
	}

	ok = true
	return
}
