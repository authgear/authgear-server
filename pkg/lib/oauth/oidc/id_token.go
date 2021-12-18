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
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
)

type UserProvider interface {
	Get(id string, role accesscontrol.Role) (*model.User, error)
}

type BaseURLProvider interface {
	BaseURL() *url.URL
}

type IDTokenIssuer struct {
	Secrets *config.OAuthKeyMaterials
	BaseURL BaseURLProvider
	Users   UserProvider
	Clock   clock.Clock
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

func (ti *IDTokenIssuer) updateTimeClaims(token jwt.Token) {
	now := ti.Clock.NowUTC()
	_ = token.Set(jwt.IssuedAtKey, now.Unix())
	_ = token.Set(jwt.ExpirationKey, now.Add(IDTokenValidDuration).Unix())
}

func (ti *IDTokenIssuer) sign(token jwt.Token) (string, error) {
	jwk, _ := ti.Secrets.Set.Get(0)
	signed, err := jwtutil.Sign(token, jwa.RS256, jwk)
	if err != nil {
		return "", err
	}
	return string(signed), nil
}

type IssueIDTokenOptions struct {
	ClientID           string
	SID                string
	Nonce              string
	AuthenticationInfo authenticationinfo.T
}

func (ti *IDTokenIssuer) IssueIDToken(opts IssueIDTokenOptions) (string, error) {
	claims := jwt.New()

	info := opts.AuthenticationInfo

	err := ti.PopulateNonPIIUserClaims(claims, info.UserID)
	if err != nil {
		return "", err
	}

	// Populate client specific claims
	_ = claims.Set(jwt.AudienceKey, opts.ClientID)

	// Populate Time specific claims
	ti.updateTimeClaims(claims)

	// Populate session specific claims
	// Note that we MUST NOT include any personal identifiable information (PII) here.
	// The ID token may be included in the GET request in form of `id_token_hint`.
	if sid := opts.SID; sid != "" {
		_ = claims.Set(string(model.ClaimSID), sid)
	}
	_ = claims.Set(string(model.ClaimAuthTime), info.AuthenticatedAt.Unix())
	if amr := info.AMR; len(amr) > 0 {
		_ = claims.Set(string(model.ClaimAMR), amr)
	}

	// Populate authorization flow specific claims
	if nonce := opts.Nonce; nonce != "" {
		_ = claims.Set("nonce", nonce)
	}

	// Sign the token.
	signed, err := ti.sign(claims)
	if err != nil {
		return "", err
	}
	return signed, nil
}

func (ti *IDTokenIssuer) VerifyIDTokenHint(client *config.OAuthClientConfig, idTokenHint string) (token jwt.Token, err error) {
	// Verify the signature.
	jwkSet, err := ti.GetPublicKeySet()
	if err != nil {
		return
	}

	_, err = jws.VerifySet([]byte(idTokenHint), jwkSet)
	if err != nil {
		return
	}
	// Parse the JWT.
	_, token, err = jwtutil.SplitWithoutVerify([]byte(idTokenHint))
	if err != nil {
		return
	}

	// Validate the claims in the JWT.
	// Here we do not use the library function jwt.Validate because
	// we do not want to validate the exp of the token.

	// We want to validate `aud` only.
	foundAud := false
	aud := client.ClientID
	for _, v := range token.Audience() {
		if v == aud {
			foundAud = true
			break
		}
	}
	if !foundAud {
		err = errors.New(`aud not satisfied`)
		return
	}

	// Normally we should also validate `iss`.
	// But `iss` can change if public_origin was changed.
	// We should still accept ID token referencing an old public_origin.

	return
}

func (ti *IDTokenIssuer) PopulateNonPIIUserClaims(token jwt.Token, userID string) error {
	user, err := ti.Users.Get(userID, config.RoleBearer)
	if err != nil {
		return err
	}

	_ = token.Set(jwt.IssuerKey, ti.Iss())
	_ = token.Set(jwt.SubjectKey, userID)
	_ = token.Set(string(model.ClaimUserIsAnonymous), user.IsAnonymous)
	_ = token.Set(string(model.ClaimUserIsVerified), user.IsVerified)
	_ = token.Set(string(model.ClaimUserCanReauthenticate), user.CanReauthenticate)

	return nil
}

func (ti *IDTokenIssuer) GetUserInfo(userID string) (map[string]interface{}, error) {
	user, err := ti.Users.Get(userID, config.RoleBearer)
	if err != nil {
		return nil, err
	}

	out := make(map[string]interface{})
	out[jwt.SubjectKey] = userID
	out[string(model.ClaimUserIsAnonymous)] = user.IsAnonymous
	out[string(model.ClaimUserIsVerified)] = user.IsVerified
	out[string(model.ClaimUserCanReauthenticate)] = user.CanReauthenticate
	for k, v := range user.StandardAttributes {
		out[k] = v
	}
	out["custom_attributes"] = user.CustomAttributes
	return out, nil
}
