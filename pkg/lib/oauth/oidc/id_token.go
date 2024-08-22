package oidc

import (
	"encoding/base64"
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
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
)

//go:generate mockgen -source=id_token.go -destination=id_token_mock_test.go -package oidc

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

	// Populate issuer.
	_ = claims.Set(jwt.IssuerKey, ti.Iss())

	err := ti.PopulateUserClaimsInIDToken(claims, info.UserID, opts.ClientLike)
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

func (ti *IDTokenIssuer) VerifyIDToken(idToken string) (token jwt.Token, err error) {
	// Verify the signature.
	jwkSet, err := ti.GetPublicKeySet()
	if err != nil {
		return
	}

	_, err = jws.Verify([]byte(idToken), jws.WithKeySet(jwkSet))
	if err != nil {
		return
	}
	// Parse the JWT.
	_, token, err = jwtutil.SplitWithoutVerify([]byte(idToken))
	if err != nil {
		return
	}

	// We used to validate `aud`.
	// However, some features like Native SSO will share a id token with multiple clients.
	// So we removed the checking of `aud`.

	// Normally we should also validate `iss`.
	// But `iss` can change if public_origin was changed.
	// We should still accept ID token referencing an old public_origin.
	// See https://linear.app/authgear/issue/DEV-1712

	return
}

func (ti *IDTokenIssuer) PopulateUserClaimsInIDToken(token jwt.Token, userID string, clientLike *oauth.ClientLike) error {
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

	_ = token.Set(jwt.SubjectKey, userID)
	_ = token.Set(string(model.ClaimUserIsAnonymous), user.IsAnonymous)
	_ = token.Set(string(model.ClaimUserIsVerified), user.IsVerified)
	_ = token.Set(string(model.ClaimUserCanReauthenticate), user.CanReauthenticate)
	_ = token.Set(string(model.ClaimAuthgearRoles), roleKeys)

	if clientLike.PIIAllowedInIDToken {
		for k, v := range user.StandardAttributes {
			isAllowed := false
			for _, scope := range clientLike.Scopes {
				if ScopeAllowsClaim(scope, k) {
					isAllowed = true
					break
				}
			}

			if isAllowed {
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

	if clientLike.IsFirstParty {
		// When the client is first party, we always include all standard attributes, all custom attributes.
		for k, v := range user.StandardAttributes {
			out[k] = v
		}

		out["custom_attributes"] = user.CustomAttributes
		out["x_web3"] = user.Web3
	} else {
		// When the client is third party, we include the standard claims according to scopes.
		for k, v := range user.StandardAttributes {
			isAllowed := false
			for _, scope := range clientLike.Scopes {
				if ScopeAllowsClaim(scope, k) {
					isAllowed = true
					break
				}
			}

			if isAllowed {
				out[k] = v
			}
		}
	}

	return out, nil
}

type IDTokenHintResolverIssuer interface {
	VerifyIDToken(idTokenHint string) (idToken jwt.Token, err error)
}

type IDTokenHintResolverSessionProvider interface {
	Get(id string) (*idpsession.IDPSession, error)
}

type IDTokenHintResolverOfflineGrantService interface {
	GetOfflineGrant(id string) (*oauth.OfflineGrant, error)
}

type IDTokenHintResolver struct {
	Issuer              IDTokenHintResolverIssuer
	Sessions            IDTokenHintResolverSessionProvider
	OfflineGrantService IDTokenHintResolverOfflineGrantService
}

func (r *IDTokenHintResolver) ResolveIDTokenHint(client *config.OAuthClientConfig, req protocol.AuthorizationRequest) (idToken jwt.Token, sidSession session.ListableSession, err error) {
	idTokenHint, ok := req.IDTokenHint()
	if !ok {
		return
	}

	idToken, err = r.Issuer.VerifyIDToken(idTokenHint)
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
		if sess, err := r.OfflineGrantService.GetOfflineGrant(sessionID); err == nil {
			sidSession = sess
		}
	default:
		panic(fmt.Errorf("oauth: unknown session type: %v", typ))
	}

	return
}
