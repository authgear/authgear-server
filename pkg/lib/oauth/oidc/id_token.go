package oidc

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

//go:generate mockgen -source=id_token.go -destination=id_token_mock_test.go -package oidc

var UserinfoScopes = []string{
	oauth.FullAccessScope,
	oauth.FullUserInfoScope,
}

var IDTokenStandardAttributes = []string{
	stdattrs.Email,
	stdattrs.EmailVerified,
	stdattrs.PhoneNumber,
	stdattrs.PhoneNumberVerified,
	stdattrs.PreferredUsername,
}

type UserProvider interface {
	Get(id string, role accesscontrol.Role) (*model.User, error)
}

type RolesAndGroupsProvider interface {
	ListEffectiveRolesByUserID(userID string) ([]*model.Role, error)
}

type BaseURLProvider interface {
	Origin() *url.URL
}

type IDTokenIssuer struct {
	Secrets        *config.OAuthKeyMaterials
	BaseURL        BaseURLProvider
	Users          UserProvider
	RolesAndGroups RolesAndGroupsProvider
	Clock          clock.Clock
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
	return ti.BaseURL.Origin().String()
}

func (ti *IDTokenIssuer) updateTimeClaims(token jwt.Token) {
	now := ti.Clock.NowUTC()
	_ = token.Set(jwt.IssuedAtKey, now.Unix())
	_ = token.Set(jwt.ExpirationKey, now.Add(IDTokenValidDuration).Unix())
}

func (ti *IDTokenIssuer) sign(token jwt.Token) (string, error) {
	jwk, _ := ti.Secrets.Set.Key(0)
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
	ClientLike         *oauth.ClientLike
	DeviceSecretHash   string
}

func (ti *IDTokenIssuer) IssueIDToken(opts IssueIDTokenOptions) (string, error) {
	claims := jwt.New()

	info := opts.AuthenticationInfo

	// For the first party client,
	// We MUST NOT include any personal identifiable information (PII) here.
	// The ID token may be included in the GET request in form of `id_token_hint`.
	nonPIIUserClaimsOnly := true
	if opts.ClientLike.PIIAllowedInIDToken {
		for _, s := range UserinfoScopes {
			if slice.ContainsString(opts.ClientLike.Scopes, s) {
				nonPIIUserClaimsOnly = false
			}
		}
	}

	err := ti.PopulateUserClaims(claims, info.UserID, nonPIIUserClaimsOnly)
	if err != nil {
		return "", err
	}

	// Populate client specific claims
	_ = claims.Set(jwt.AudienceKey, opts.ClientID)

	// Populate Time specific claims
	ti.updateTimeClaims(claims)

	// Populate session specific claims
	if sid := opts.SID; sid != "" {
		_ = claims.Set(string(model.ClaimSID), sid)
	}
	_ = claims.Set(string(model.ClaimAuthTime), info.AuthenticatedAt.Unix())
	if amr := info.AMR; len(amr) > 0 {
		_ = claims.Set(string(model.ClaimAMR), amr)
	}
	if dshash := opts.DeviceSecretHash; dshash != "" {
		_ = claims.Set(string(model.ClaimDeviceSecretHash), dshash)
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

func (ti *IDTokenIssuer) VerifyIDTokenHintWithoutClient(idTokenHint string) (token jwt.Token, err error) {
	// Verify the signature.
	jwkSet, err := ti.GetPublicKeySet()
	if err != nil {
		return
	}

	_, err = jws.Verify([]byte(idTokenHint), jws.WithKeySet(jwkSet))
	if err != nil {
		return
	}
	// Parse the JWT.
	_, token, err = jwtutil.SplitWithoutVerify([]byte(idTokenHint))
	if err != nil {
		return
	}

	return
}

func (ti *IDTokenIssuer) VerifyIDTokenHint(client *config.OAuthClientConfig, idTokenHint string) (token jwt.Token, err error) {
	token, err = ti.VerifyIDTokenHintWithoutClient(idTokenHint)
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
	return ti.PopulateUserClaims(token, userID, true)
}

func (ti *IDTokenIssuer) PopulateUserClaims(token jwt.Token, userID string, nonPIIUserClaimsOnly bool) error {
	user, err := ti.Users.Get(userID, config.RoleBearer)
	if err != nil {
		return err
	}

	roles, err := ti.RolesAndGroups.ListEffectiveRolesByUserID(userID)
	if err != nil {
		return err
	}
	roleKeys := make([]string, len(roles))
	for i := range roles {
		roleKeys[i] = roles[i].Key
	}

	_ = token.Set(jwt.IssuerKey, ti.Iss())
	_ = token.Set(jwt.SubjectKey, userID)
	_ = token.Set(string(model.ClaimUserIsAnonymous), user.IsAnonymous)
	_ = token.Set(string(model.ClaimUserIsVerified), user.IsVerified)
	_ = token.Set(string(model.ClaimUserCanReauthenticate), user.CanReauthenticate)
	_ = token.Set(string(model.ClaimAuthgearRoles), roleKeys)

	if !nonPIIUserClaimsOnly {
		for k, v := range user.StandardAttributes {
			if slice.ContainsString(IDTokenStandardAttributes, k) {
				_ = token.Set(k, v)
			}
		}
	}

	return nil
}

func (ti *IDTokenIssuer) GetUserInfo(userID string, clientLike *oauth.ClientLike) (map[string]interface{}, error) {
	user, err := ti.Users.Get(userID, config.RoleBearer)
	if err != nil {
		return nil, err
	}

	roles, err := ti.RolesAndGroups.ListEffectiveRolesByUserID(userID)
	if err != nil {
		return nil, err
	}
	roleKeys := make([]string, len(roles))
	for i := range roles {
		roleKeys[i] = roles[i].Key
	}

	out := make(map[string]interface{})
	out[jwt.SubjectKey] = userID
	out[string(model.ClaimUserIsAnonymous)] = user.IsAnonymous
	out[string(model.ClaimUserIsVerified)] = user.IsVerified
	out[string(model.ClaimUserCanReauthenticate)] = user.CanReauthenticate
	out[string(model.ClaimAuthgearRoles)] = roleKeys

	nonPIIUserClaimsOnly := true
	// When the client is first party
	// always include userinfo for the userinfo endpoint
	// We check the scopes only for third party client
	if clientLike.IsFirstParty {
		nonPIIUserClaimsOnly = false
	} else {
		for _, s := range UserinfoScopes {
			if slice.ContainsString(clientLike.Scopes, s) {
				nonPIIUserClaimsOnly = false
			}
		}
	}
	if nonPIIUserClaimsOnly {
		return out, nil
	}

	// todo add role
	// Populate userinfo claims
	for k, v := range user.StandardAttributes {
		out[k] = v
	}
	out["custom_attributes"] = user.CustomAttributes
	out["x_web3"] = user.Web3
	return out, nil
}

type IDTokenHintResolverIssuer interface {
	VerifyIDTokenHint(client *config.OAuthClientConfig, idTokenHint string) (idToken jwt.Token, err error)
}

type IDTokenHintResolverSessionProvider interface {
	Get(id string) (*idpsession.IDPSession, error)
}

type IDTokenHintResolver struct {
	Issuer        IDTokenHintResolverIssuer
	Sessions      IDTokenHintResolverSessionProvider
	OfflineGrants oauth.OfflineGrantStore
}

func (r *IDTokenHintResolver) ResolveIDTokenHint(client *config.OAuthClientConfig, req protocol.AuthorizationRequest) (idToken jwt.Token, sidSession session.ListableSession, err error) {
	idTokenHint, ok := req.IDTokenHint()
	if !ok {
		return
	}

	idToken, err = r.Issuer.VerifyIDTokenHint(client, idTokenHint)
	if err != nil {
		return
	}

	sidInterface, ok := idToken.Get(string(model.ClaimSID))
	if !ok {
		return
	}

	sid, ok := sidInterface.(string)
	if !ok {
		return
	}

	typ, sessionID, ok := DecodeSID(sid)
	if !ok {
		return
	}

	switch typ {
	case session.TypeIdentityProvider:
		if sess, err := r.Sessions.Get(sessionID); err == nil {
			sidSession = sess
		}
	case session.TypeOfflineGrant:
		if sess, err := r.OfflineGrants.GetOfflineGrant(sessionID); err == nil {
			sidSession = sess
		}
	default:
		panic(fmt.Errorf("oauth: unknown session type: %v", typ))
	}

	return
}
