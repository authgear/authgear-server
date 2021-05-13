package oidc

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/jwsutil"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
)

type IDPSessionService interface {
	Get(id string) (*idpsession.IDPSession, error)
}

type UserProvider interface {
	Get(id string) (*model.User, error)
}

type BaseURLProvider interface {
	BaseURL() *url.URL
}

type IDTokenIssuer struct {
	IDPSessions   IDPSessionService
	OfflineGrants oauth.OfflineGrantStore
	Secrets       *config.OAuthKeyMaterials
	BaseURL       BaseURLProvider
	Users         UserProvider
	Clock         clock.Clock
}

// IDTokenValidDuration is the valid period of ID token.
// It can be short, since id_token_hint should accept expired ID tokens.
const IDTokenValidDuration = duration.Short

type SessionLike interface {
	SessionID() string
	SessionType() session.Type
}

func EncodeSID(s SessionLike) string {
	raw := fmt.Sprintf("%s:%s", s.SessionType(), s.SessionID())
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

func (ti *IDTokenIssuer) GetPublicKeySet() (jwk.Set, error) {
	return jwk.PublicSetOf(ti.Secrets.Set)
}

func (ti *IDTokenIssuer) Iss() string {
	return ti.BaseURL.BaseURL().String()
}

func (ti *IDTokenIssuer) ValidateIDTokenHint(idToken string) (session.Session, error) {
	compact := []byte(idToken)
	keySet, err := ti.GetPublicKeySet()
	if err != nil {
		return nil, err
	}
	_, payload, err := jwsutil.VerifyWithSet(keySet, compact)
	if err != nil {
		return nil, err
	}

	// https://openid.net/specs/openid-connect-rpinitiated-1_0.html#RPLogout

	// When an id_token_hint parameter is present, the OP MUST validate that it was the issuer of the ID Token.
	if payload.Issuer() != ti.Iss() {
		return nil, errors.New("iss not satisfied")
	}

	// The OP SHOULD accept ID Tokens when the RP identified by the ID Token's aud claim and/or sid claim has a current session or had a recent session at the OP,
	// even when the exp time has passed.

	// If the ID Token's sid claim does not correspond to the RP's current session or a recent session at the OP,
	// the OP SHOULD treat the logout request as suspect, and MAY decline to act upon it.
	sidIface, ok := payload.Get("sid")
	if !ok {
		return nil, errors.New("missing sid")
	}
	sid, ok := sidIface.(string)
	if !ok {
		return nil, fmt.Errorf("expected sid to be string but was %T", sidIface)
	}

	typ, sessionID, ok := DecodeSID(sid)
	if !ok {
		return nil, fmt.Errorf("invalid sid")
	}

	switch typ {
	case session.TypeIdentityProvider:
		var s session.Session
		s, err = ti.IDPSessions.Get(sessionID)
		if err != nil {
			return nil, err
		}
		return s, nil
	case session.TypeOfflineGrant:
		// TODO(logout): Support id_token_hint with sid referring to a offline grant.
		// 1. Check if the cookie session and the offline grant is referring to the same user.
		// 2. Check if the offline grant has full access scope.
		return nil, errors.New("id_token_hint does not support offline grant yet")
	default:
		panic(fmt.Errorf("unexpected session type: %s", typ))
	}
}

func (ti *IDTokenIssuer) IssueIDToken(client *config.OAuthClientConfig, s session.Session, nonce string) (string, error) {
	claims, err := ti.LoadUserClaims(s.SessionAttrs().UserID)
	if err != nil {
		return "", err
	}

	now := ti.Clock.NowUTC()

	_ = claims.Set(jwt.AudienceKey, client.ClientID)

	_ = claims.Set("sid", EncodeSID(s))

	_ = claims.Set(jwt.IssuedAtKey, now.Unix())
	_ = claims.Set(jwt.ExpirationKey, now.Add(IDTokenValidDuration).Unix())

	// Avoid putting user claims in ID token
	// ID token is used as `id_token_hint` in `end_session_endpoint` for direct logout
	// To avoid putting user data in `GET` request parameters, so those claims
	// are removed from the ID token

	if nonce != "" {
		_ = claims.Set("nonce", nonce)
	}

	jwk, _ := ti.Secrets.Set.Get(0)

	signed, err := jwtutil.Sign(claims, jwa.RS256, jwk)
	if err != nil {
		return "", err
	}

	return string(signed), nil
}

func (ti *IDTokenIssuer) LoadUserClaims(userID string) (jwt.Token, error) {
	user, err := ti.Users.Get(userID)
	if err != nil {
		return nil, err
	}

	claims := jwt.New()
	_ = claims.Set(jwt.IssuerKey, ti.Iss())
	_ = claims.Set(jwt.SubjectKey, userID)
	_ = claims.Set(string(authn.ClaimUserIsAnonymous), user.IsAnonymous)
	_ = claims.Set(string(authn.ClaimUserIsVerified), user.IsVerified)

	// Avoid putting user claims in LoadUserClaims
	// LoadUserClaims defines claims in ID token and
	// ID token is used as `id_token_hint` in `end_session_endpoint` for direct logout
	// To avoid putting user data in `GET` request parameters, so don't
	// put user claims in LoadUserClaims

	return claims, nil
}
