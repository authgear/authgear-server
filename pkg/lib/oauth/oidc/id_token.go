package oidc

import (
	"context"
	"fmt"
	"net/url"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/lib/userinfo"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
)

//go:generate go tool mockgen -source=id_token.go -destination=id_token_mock_test.go -package oidc

type UserInfoService interface {
	GetUserInfoBearer(ctx context.Context, userID string) (*userinfo.UserInfo, error)
}

type BaseURLProvider interface {
	Origin() *url.URL
}

type IDTokenIssuerIdentityService interface {
	ListIdentitiesThatHaveStandardAttributes(ctx context.Context, userID string) ([]*identity.Info, error)
}

type IDTokenIssuerEventService interface {
	PrepareBlockingEventWithTx(ctx context.Context, payload event.BlockingPayload) (e *event.Event, err error)
	DispatchEventWithoutTx(ctx context.Context, e *event.Event) (err error)
}

type IDTokenIssuer struct {
	Secrets         *config.OAuthKeyMaterials
	BaseURL         BaseURLProvider
	UserInfoService UserInfoService
	Events          IDTokenIssuerEventService
	Identities      IDTokenIssuerIdentityService
	Clock           clock.Clock
}

// IDTokenValidDuration is the valid period of ID token.
// It can be short, since id_token_hint should accept expired ID tokens.
const IDTokenValidDuration = duration.Short

func (ti *IDTokenIssuer) GetPublicKeySet() (jwk.Set, error) {
	return jwk.PublicSetOf(ti.Secrets.Set)
}

func (ti *IDTokenIssuer) Iss() string {
	return ti.BaseURL.Origin().String()
}

func (ti *IDTokenIssuer) sign(token jwt.Token) (string, error) {
	jwk, _ := ti.Secrets.Set.Key(0)
	signed, err := jwtutil.Sign(token, jwa.RS256, jwk)
	if err != nil {
		return "", err
	}
	return string(signed), nil
}

type PrepareIDTokenOptions struct {
	ClientID           string
	SID                string
	Nonce              string
	AuthenticationInfo authenticationinfo.T
	ClientLike         *oauth.ClientLike
	DeviceSecretHash   string
	IdentitySpecs      []*identity.Spec
}

type PrepareIDTokenResult struct {
	event     *event.Event
	forBackup map[string]interface{}
}

func (ti *IDTokenIssuer) PrepareIDToken(ctx context.Context, opts PrepareIDTokenOptions) (*PrepareIDTokenResult, error) {
	claims := jwt.New()

	info := opts.AuthenticationInfo

	// iss
	_ = claims.Set(jwt.IssuerKey, ti.Iss())
	// aud
	_ = claims.Set(jwt.AudienceKey, opts.ClientID)
	now := ti.Clock.NowUTC()
	// iat
	_ = claims.Set(jwt.IssuedAtKey, now.Unix())
	// exp
	_ = claims.Set(jwt.ExpirationKey, now.Add(IDTokenValidDuration).Unix())
	// auth_time
	_ = claims.Set(string(model.ClaimAuthTime), info.AuthenticatedAt.Unix())
	// sid
	if sid := opts.SID; sid != "" {
		_ = claims.Set(string(model.ClaimSID), sid)
	}
	// amr
	if amr := info.AMR; len(amr) > 0 {
		_ = claims.Set(string(model.ClaimAMR), amr)
	}
	// ds_hash
	if dshash := opts.DeviceSecretHash; dshash != "" {
		_ = claims.Set(string(model.ClaimDeviceSecretHash), dshash)
	}
	// nonce
	if nonce := opts.Nonce; nonce != "" {
		_ = claims.Set("nonce", nonce)
	}

	err := ti.PopulateUserClaimsInIDToken(ctx, claims, info.UserID, opts.ClientLike)
	if err != nil {
		return nil, err
	}

	// https://authgear.com/claims/oauth/asserted
	var oauthUsed []map[string]any
	for _, idenSpec := range opts.IdentitySpecs {
		if idenSpec.Type == model.IdentityTypeOAuth && idenSpec.OAuth != nil && idenSpec.OAuth.IncludeIdentityAttributesInIDToken {
			oauthUsed = append(oauthUsed, idenSpec.OAuth.ToClaimsForIDToken())
		}
	}
	if len(oauthUsed) > 0 {
		_ = claims.Set(string(model.ClaimOAuthAsserted), oauthUsed)
	}

	forMutation, forBackup, err := jwtutil.PrepareForMutations(claims)
	if err != nil {
		return nil, err
	}

	identities, err := ti.Identities.ListIdentitiesThatHaveStandardAttributes(ctx, opts.AuthenticationInfo.UserID)
	if err != nil {
		return nil, err
	}

	var identityModels []model.Identity
	for _, i := range identities {
		identityModels = append(identityModels, i.ToModel())
	}

	eventPayload := &blocking.OIDCIDTokenPreCreateBlockingEventPayload{
		UserRef: model.UserRef{
			Meta: model.Meta{
				ID: opts.AuthenticationInfo.UserID,
			},
		},
		Identities: identityModels,
		JWT: blocking.OIDCIDToken{
			Payload: forMutation,
		},
	}

	event, err := ti.Events.PrepareBlockingEventWithTx(ctx, eventPayload)
	if err != nil {
		return nil, err
	}

	return &PrepareIDTokenResult{
		event:     event,
		forBackup: forBackup,
	}, nil
}

type MakeIDTokenFromPreparationResultOptions struct {
	PreparationResult *PrepareIDTokenResult
}

func (ti *IDTokenIssuer) MakeIDTokenFromPreparationResult(
	ctx context.Context,
	options MakeIDTokenFromPreparationResultOptions,
) (string, error) {
	err := ti.Events.DispatchEventWithoutTx(ctx, options.PreparationResult.event)
	if err != nil {
		return "", err
	}

	eventPayload := options.PreparationResult.event.Payload.(*blocking.OIDCIDTokenPreCreateBlockingEventPayload)

	claims, err := jwtutil.ApplyMutations(
		eventPayload.JWT.Payload,
		options.PreparationResult.forBackup,
	)
	if err != nil {
		return "", err
	}

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

func (ti *IDTokenIssuer) PopulateUserClaimsInIDToken(ctx context.Context, token jwt.Token, userID string, clientLike *oauth.ClientLike) error {
	userInfo, err := ti.UserInfoService.GetUserInfoBearer(ctx, userID)
	if err != nil {
		return err
	}

	_ = token.Set(jwt.SubjectKey, userID)
	_ = token.Set(string(model.ClaimUserIsAnonymous), userInfo.User.IsAnonymous)
	_ = token.Set(string(model.ClaimUserIsVerified), userInfo.User.IsVerified)
	_ = token.Set(string(model.ClaimUserCanReauthenticate), userInfo.User.CanReauthenticate)
	_ = token.Set(string(model.ClaimAuthgearRoles), userInfo.EffectiveRoleKeys)

	if clientLike.PIIAllowedInIDToken {
		for k, v := range userInfo.User.StandardAttributes {
			isAllowed := false
			for _, scope := range clientLike.Scopes {
				if oauth.ScopeAllowsClaim(scope, k) {
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

func (ti *IDTokenIssuer) GetUserInfo(ctx context.Context, userID string, clientLike *oauth.ClientLike) (map[string]interface{}, error) {
	userInfo, err := ti.UserInfoService.GetUserInfoBearer(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make(map[string]interface{})
	out[jwt.SubjectKey] = userID
	out[string(model.ClaimUserIsAnonymous)] = userInfo.User.IsAnonymous
	out[string(model.ClaimUserIsVerified)] = userInfo.User.IsVerified
	out[string(model.ClaimUserCanReauthenticate)] = userInfo.User.CanReauthenticate
	out[string(model.ClaimAuthgearRoles)] = userInfo.EffectiveRoleKeys

	if clientLike.IsFirstParty {
		// When the client is first party, we always include all standard attributes, all custom attributes.
		for k, v := range userInfo.User.StandardAttributes {
			out[k] = v
		}

		out["custom_attributes"] = userInfo.User.CustomAttributes
		out["x_web3"] = userInfo.User.Web3
	} else {
		// When the client is third party, we include the standard claims according to scopes.
		for k, v := range userInfo.User.StandardAttributes {
			isAllowed := false
			for _, scope := range clientLike.Scopes {
				if oauth.ScopeAllowsClaim(scope, k) {
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
	Get(ctx context.Context, id string) (*idpsession.IDPSession, error)
}

type IDTokenHintResolverOfflineGrantService interface {
	GetOfflineGrant(ctx context.Context, id string) (*oauth.OfflineGrant, error)
}

type IDTokenHintResolver struct {
	Issuer              IDTokenHintResolverIssuer
	Sessions            IDTokenHintResolverSessionProvider
	OfflineGrantService IDTokenHintResolverOfflineGrantService
}

func (r *IDTokenHintResolver) ResolveIDTokenHint(ctx context.Context, client *config.OAuthClientConfig, req protocol.AuthorizationRequest) (idToken jwt.Token, sidSession session.ListableSession, err error) {
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

	typ, sessionID, ok := oauth.DecodeSID(sid)
	if !ok {
		return
	}

	switch typ {
	case session.TypeIdentityProvider:
		if sess, err := r.Sessions.Get(ctx, sessionID); err == nil {
			sidSession = sess
		}
	case session.TypeOfflineGrant:
		if sess, err := r.OfflineGrantService.GetOfflineGrant(ctx, sessionID); err == nil {
			sidSession = sess
		}
	default:
		panic(fmt.Errorf("oauth: unknown session type: %v", typ))
	}

	return
}
