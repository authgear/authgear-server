package handler

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	identitybiometric "github.com/authgear/authgear-server/pkg/lib/authn/identity/biometric"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	interactionintents "github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/jwkutil"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

const AnonymousRequestGrantType = "urn:authgear:params:oauth:grant-type:anonymous-request"
const BiometricRequestGrantType = "urn:authgear:params:oauth:grant-type:biometric-request"

// nolint: gosec
const IDTokenGrantType = "urn:authgear:params:oauth:grant-type:id-token"

const AppSessionTokenDuration = duration.Short

// whitelistedGrantTypes is a list of grant types that would be always allowed
// to all clients.
var whitelistedGrantTypes = []string{
	AnonymousRequestGrantType,
	BiometricRequestGrantType,
	IDTokenGrantType,
}

type IDTokenIssuer interface {
	IssueIDToken(opts oidc.IssueIDTokenOptions) (token string, err error)
}

type AccessTokenIssuer interface {
	EncodeAccessToken(client *config.OAuthClientConfig, grant *oauth.AccessGrant, userID string, token string) (string, error)
}

type TokenHandlerUserFacade interface {
	GetRaw(id string) (*user.User, error)
}

type TokenHandlerLogger struct{ *log.Logger }

func NewTokenHandlerLogger(lf *log.Factory) TokenHandlerLogger {
	return TokenHandlerLogger{lf.New("oauth-token")}
}

type TokenHandler struct {
	AppID                  config.AppID
	Config                 *config.OAuthConfig
	IdentityFeatureConfig  *config.IdentityFeatureConfig
	OAuthClientCredentials *config.OAuthClientCredentials
	Logger                 TokenHandlerLogger

	Authorizations      oauth.AuthorizationStore
	CodeGrants          oauth.CodeGrantStore
	OfflineGrants       oauth.OfflineGrantStore
	AppSessionTokens    oauth.AppSessionTokenStore
	OfflineGrantService oauth.OfflineGrantService
	Graphs              GraphService
	IDTokenIssuer       IDTokenIssuer
	Clock               clock.Clock
	TokenService        TokenService
}

func (h *TokenHandler) Handle(rw http.ResponseWriter, req *http.Request, r protocol.TokenRequest) httputil.Result {
	client := resolveClient(h.Config, r)
	if client == nil {
		return tokenResultError{
			Response: protocol.NewErrorResponse("invalid_client", "invalid client ID"),
		}
	}

	result, err := h.doHandle(rw, req, client, r)
	if err != nil {
		var oauthError *protocol.OAuthProtocolError
		resultErr := tokenResultError{}
		if errors.As(err, &oauthError) {
			resultErr.StatusCode = oauthError.StatusCode
			resultErr.Response = oauthError.Response
		} else {
			h.Logger.WithError(err).Error("authz handler failed")
			resultErr.Response = protocol.NewErrorResponse("server_error", "internal server error")
			resultErr.InternalError = true
		}
		result = resultErr
	}

	return result
}

func (h *TokenHandler) doHandle(
	rw http.ResponseWriter,
	req *http.Request,
	client *config.OAuthClientConfig,
	r protocol.TokenRequest,
) (httputil.Result, error) {
	if err := h.validateRequest(r, client); err != nil {
		return nil, err
	}

	allowedGrantTypes := client.GrantTypes
	if len(allowedGrantTypes) == 0 {
		allowedGrantTypes = []string{"authorization_code"}
	}
	allowedGrantTypes = append(allowedGrantTypes, whitelistedGrantTypes...)

	ok := false
	for _, grantType := range allowedGrantTypes {
		if r.GrantType() == grantType {
			ok = true
			break
		}
	}
	if !ok {
		return nil, protocol.NewError("unauthorized_client", "grant type is not allowed for this client")
	}

	switch r.GrantType() {
	case "authorization_code":
		return h.handleAuthorizationCode(client, r)
	case "refresh_token":
		resp, err := h.handleRefreshToken(client, r)
		if err != nil {
			return nil, err
		}
		return tokenResultOK{Response: resp}, nil
	case AnonymousRequestGrantType:
		return h.handleAnonymousRequest(client, r)
	case BiometricRequestGrantType:
		return h.handleBiometricRequest(rw, req, client, r)
	case IDTokenGrantType:
		return h.handleIDToken(rw, req, client, r)
	default:
		panic("oauth: unexpected grant type")
	}
}

func (h *TokenHandler) validateRequest(r protocol.TokenRequest, client *config.OAuthClientConfig) error {
	switch r.GrantType() {
	case "authorization_code":
		if r.Code() == "" {
			return protocol.NewError("invalid_request", "code is required")
		}
		if client.ClientParty() == config.ClientPartyFirst {
			if r.CodeVerifier() == "" {
				return protocol.NewError("invalid_request", "PKCE code verifier is required")
			}
		} else {
			if r.ClientSecret() == "" {
				return protocol.NewError("invalid_request", "client secret is required")
			}
		}
	case "refresh_token":
		if r.RefreshToken() == "" {
			return protocol.NewError("invalid_request", "refresh token is required")
		}
	case AnonymousRequestGrantType:
		if r.JWT() == "" {
			return protocol.NewError("invalid_request", "jwt is required")
		}
	case BiometricRequestGrantType:
		if r.JWT() == "" {
			return protocol.NewError("invalid_request", "jwt is required")
		}
	case IDTokenGrantType:
		break
	default:
		return protocol.NewError("unsupported_grant_type", "grant type is not supported")
	}

	return nil
}

var errInvalidAuthzCode = protocol.NewError("invalid_grant", "invalid authorization code")

func (h *TokenHandler) handleAuthorizationCode(
	client *config.OAuthClientConfig,
	r protocol.TokenRequest,
) (httputil.Result, error) {
	deviceInfo, err := r.DeviceInfo()
	if err != nil {
		return nil, protocol.NewError("invalid_request", err.Error())
	}

	codeHash := oauth.HashToken(r.Code())
	codeGrant, err := h.CodeGrants.GetCodeGrant(codeHash)
	if errors.Is(err, oauth.ErrGrantNotFound) {
		return nil, errInvalidAuthzCode
	} else if err != nil {
		return nil, err
	}

	if h.Clock.NowUTC().After(codeGrant.ExpireAt) {
		return nil, errInvalidAuthzCode
	}

	if codeGrant.RedirectURI != r.RedirectURI() {
		return nil, protocol.NewError("invalid_request", "invalid redirect URI")
	}

	// verify pkce
	needVerifyPKCE := client.ClientParty() == config.ClientPartyFirst || codeGrant.PKCEChallenge != "" || r.CodeVerifier() != ""
	if needVerifyPKCE {
		if codeGrant.PKCEChallenge == "" || r.CodeVerifier() == "" || !verifyPKCE(codeGrant.PKCEChallenge, r.CodeVerifier()) {
			return nil, errInvalidAuthzCode
		}
	}

	// verify client secret
	needClientSecret := client.ClientParty() == config.ClientPartyThird
	if needClientSecret {
		if r.ClientSecret() == "" {
			return nil, protocol.NewError("invalid_request", "invalid client secret")
		}
		credentialsItem, ok := h.OAuthClientCredentials.Lookup(client.ClientID)
		if !ok {
			return nil, protocol.NewError("invalid_request", "client secret is not supported for the client")
		}

		pass := false
		keys, _ := jwkutil.ExtractOctetKeys(credentialsItem.Set)
		for _, clientSecret := range keys {
			if subtle.ConstantTimeCompare([]byte(r.ClientSecret()), clientSecret) == 1 {
				pass = true
			}
		}
		if !pass {
			return nil, protocol.NewError("invalid_request", "invalid client secret")
		}
	}

	authz, err := h.Authorizations.GetByID(codeGrant.AuthorizationID)
	if errors.Is(err, oauth.ErrAuthorizationNotFound) {
		return nil, errInvalidAuthzCode
	} else if err != nil {
		return nil, err
	}

	resp, err := h.issueTokensForAuthorizationCode(client, codeGrant, authz, deviceInfo)
	if err != nil {
		return nil, err
	}

	err = h.CodeGrants.DeleteCodeGrant(codeGrant)
	if err != nil {
		h.Logger.WithError(err).Error("failed to invalidate code grant")
	}

	return tokenResultOK{Response: resp}, nil
}

func (h *TokenHandler) handleRefreshToken(
	client *config.OAuthClientConfig,
	r protocol.TokenRequest,
) (protocol.TokenResponse, error) {
	deviceInfo, err := r.DeviceInfo()
	if err != nil {
		return nil, protocol.NewError("invalid_request", err.Error())
	}

	authz, offlineGrant, err := h.TokenService.ParseRefreshToken(r.RefreshToken())
	if err != nil {
		return nil, err
	}

	resp, err := h.issueTokensForRefreshToken(client, offlineGrant, authz)
	if err != nil {
		return nil, err
	}

	if client.ClientID != offlineGrant.ClientID {
		return nil, protocol.NewError("invalid_request", "client id doesn't match the refresh token")
	}

	expiry, err := h.OfflineGrantService.ComputeOfflineGrantExpiry(offlineGrant)
	if err != nil {
		return nil, err
	}
	_, err = h.OfflineGrants.UpdateOfflineGrantDeviceInfo(offlineGrant.ID, deviceInfo, expiry)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type anonymousTokenInput struct {
	JWT string
}

func (i *anonymousTokenInput) GetAnonymousRequestToken() string {
	return i.JWT
}

func (i *anonymousTokenInput) SignUpAnonymousUserWithoutKey() bool {
	return false
}

func (i *anonymousTokenInput) GetPromoteUserAndIdentityID() (string, string) { return "", "" }

var _ nodes.InputUseIdentityAnonymous = &anonymousTokenInput{}

func (h *TokenHandler) handleAnonymousRequest(
	client *config.OAuthClientConfig,
	r protocol.TokenRequest,
) (httputil.Result, error) {
	if client.ClientParty() == config.ClientPartyThird {
		return nil, protocol.NewError(
			"unauthorized_client",
			"third-party clients may not use anonymous user",
		)
	}

	deviceInfo, err := r.DeviceInfo()
	if err != nil {
		return nil, protocol.NewError("invalid_request", err.Error())
	}

	var graph *interaction.Graph
	err = h.Graphs.DryRun("", func(ctx *interaction.Context) (*interaction.Graph, error) {
		var err error
		intent := &interactionintents.IntentAuthenticate{
			Kind:                     interactionintents.IntentAuthenticateKindLogin,
			SuppressIDPSessionCookie: true,
		}
		graph, err = h.Graphs.NewGraph(ctx, intent)
		if err != nil {
			return nil, err
		}

		var edges []interaction.Edge
		graph, edges, err = h.Graphs.Accept(ctx, graph, &anonymousTokenInput{
			JWT: r.JWT(),
		})
		if len(edges) != 0 {
			return nil, errors.New("interaction not completed for anonymous users")
		} else if err != nil {
			return nil, err
		}

		return graph, nil
	})

	if apierrors.IsKind(err, interaction.InvariantViolated) &&
		apierrors.AsAPIError(err).HasCause("AnonymousUserDisallowed") {
		return nil, protocol.NewError("unauthorized_client", "AnonymousUserDisallowed")
	} else if errors.Is(err, interaction.ErrInvalidCredentials) {
		return nil, protocol.NewError("invalid_grant", interaction.InvalidCredentials.Reason)
	} else if err != nil {
		return nil, err
	}

	info := authenticationinfo.T{
		UserID:          graph.MustGetUserID(),
		AuthenticatedAt: h.Clock.NowUTC(),
	}

	err = h.Graphs.Run("", graph)
	if apierrors.IsAPIError(err) {
		return nil, protocol.NewError("invalid_request", err.Error())
	} else if err != nil {
		return nil, err
	}

	// TODO(oauth): allow specifying scopes
	scopes := []string{"openid", oauth.FullAccessScope}

	authz, err := checkAndGrantAuthorization(
		h.Authorizations,
		h.Clock.NowUTC(),
		h.AppID,
		client.ClientID,
		info.UserID,
		scopes,
	)
	if err != nil {
		return nil, err
	}

	resp := protocol.TokenResponse{}

	// SSOEnabled is false for refresh tokens that are granted by anonymous login
	opts := IssueOfflineGrantOptions{
		Scopes:             scopes,
		AuthorizationID:    authz.ID,
		AuthenticationInfo: info,
		DeviceInfo:         deviceInfo,
		SSOEnabled:         false,
	}
	offlineGrant, err := h.TokenService.IssueOfflineGrant(client, opts, resp)
	if err != nil {
		return nil, err
	}

	err = h.TokenService.IssueAccessGrant(client, scopes, authz.ID, authz.UserID,
		offlineGrant.ID, oauth.GrantSessionKindOffline, resp)
	if err != nil {
		return nil, err
	}

	return tokenResultOK{Response: resp}, nil
}

func (h *TokenHandler) handleBiometricRequest(
	rw http.ResponseWriter,
	req *http.Request,
	client *config.OAuthClientConfig,
	r protocol.TokenRequest,
) (httputil.Result, error) {
	if *h.IdentityFeatureConfig.Biometric.Disabled {
		return nil, protocol.NewError(
			"invalid_request",
			"biometric authentication is disabled",
		)
	}

	if client.ClientParty() == config.ClientPartyThird {
		return nil, protocol.NewError(
			"unauthorized_client",
			"third-party clients may not use biometric authentication",
		)
	}
	_, payload, err := jwtutil.SplitWithoutVerify([]byte(r.JWT()))
	if err != nil {
		return nil, protocol.NewError("invalid_request", err.Error())
	}
	actionIface, _ := payload.Get("action")
	action, _ := actionIface.(string)
	switch action {
	case string(identitybiometric.RequestActionSetup):
		return h.handleBiometricSetup(req, client, r)
	case string(identitybiometric.RequestActionAuthenticate):
		return h.handleBiometricAuthenticate(client, r)
	default:
		return nil, protocol.NewError("invalid_request", fmt.Sprintf("invalid action: %v", actionIface))
	}
}

type biometricInput struct {
	JWT string
}

func (i *biometricInput) GetBiometricRequestToken() string {
	return i.JWT
}

func (h *TokenHandler) handleBiometricSetup(
	req *http.Request,
	client *config.OAuthClientConfig,
	r protocol.TokenRequest,
) (httputil.Result, error) {
	s := session.GetSession(req.Context())
	if s == nil {
		return nil, protocol.NewErrorStatusCode("invalid_request", "biometric setup requires authenticated user", http.StatusUnauthorized)
	}

	var graph *interaction.Graph
	err := h.Graphs.DryRun("", func(ctx *interaction.Context) (*interaction.Graph, error) {
		var err error
		graph, err = h.Graphs.NewGraph(ctx, interactionintents.NewIntentAddIdentity(s.GetAuthenticationInfo().UserID))
		if err != nil {
			return nil, err
		}

		var edges []interaction.Edge
		graph, edges, err = h.Graphs.Accept(ctx, graph, &biometricInput{
			JWT: r.JWT(),
		})
		if len(edges) != 0 {
			return nil, errors.New("interaction no completed for biometric setup")
		} else if err != nil {
			return nil, err
		}

		return graph, nil
	})

	if apierrors.IsKind(err, interaction.InvariantViolated) &&
		apierrors.AsAPIError(err).HasCause("BiometricDisallowed") {
		return nil, protocol.NewError("unauthorized_client", "BiometricDisallowed")
	} else if apierrors.IsKind(err, interaction.InvariantViolated) &&
		apierrors.AsAPIError(err).HasCause("AnonymousUserAddIdentity") {
		return nil, protocol.NewError("unauthorized_client", "AnonymousUserAddIdentity")
	} else if errors.Is(err, interaction.ErrInvalidCredentials) {
		return nil, protocol.NewError("invalid_grant", interaction.InvalidCredentials.Reason)
	} else if err != nil {
		return nil, err
	}

	err = h.Graphs.Run("", graph)
	if apierrors.IsAPIError(err) {
		return nil, protocol.NewError("invalid_request", err.Error())
	} else if err != nil {
		return nil, err
	}

	return tokenResultEmpty{}, nil
}

func (h *TokenHandler) handleBiometricAuthenticate(
	client *config.OAuthClientConfig,
	r protocol.TokenRequest,
) (httputil.Result, error) {
	deviceInfo, err := r.DeviceInfo()
	if err != nil {
		return nil, protocol.NewError("invalid_request", err.Error())
	}

	var graph *interaction.Graph
	err = h.Graphs.DryRun("", func(ctx *interaction.Context) (*interaction.Graph, error) {
		var err error
		intent := &interactionintents.IntentAuthenticate{
			Kind:                     interactionintents.IntentAuthenticateKindLogin,
			SuppressIDPSessionCookie: true,
		}
		graph, err = h.Graphs.NewGraph(ctx, intent)
		if err != nil {
			return nil, err
		}

		var edges []interaction.Edge
		graph, edges, err = h.Graphs.Accept(ctx, graph, &biometricInput{
			JWT: r.JWT(),
		})
		if len(edges) != 0 {
			return nil, errors.New("interaction no completed for biometric authenticate")
		} else if err != nil {
			return nil, err
		}

		return graph, nil
	})

	if apierrors.IsKind(err, interaction.InvariantViolated) &&
		apierrors.AsAPIError(err).HasCause("BiometricDisallowed") {
		return nil, protocol.NewError("unauthorized_client", "BiometricDisallowed")
	} else if errors.Is(err, interaction.ErrInvalidCredentials) {
		return nil, protocol.NewError("invalid_grant", interaction.InvalidCredentials.Reason)
	} else if err != nil {
		return nil, err
	}

	info := authenticationinfo.T{
		UserID:          graph.MustGetUserID(),
		AMR:             graph.GetAMR(),
		AuthenticatedAt: h.Clock.NowUTC(),
	}
	biometricIdentity := graph.MustGetUserLastIdentity()

	err = h.Graphs.Run("", graph)
	if apierrors.IsAPIError(err) {
		return nil, protocol.NewError("invalid_request", err.Error())
	} else if err != nil {
		return nil, err
	}

	// TODO(oauth): allow specifying scopes
	scopes := []string{"openid", oauth.FullAccessScope}

	authz, err := checkAndGrantAuthorization(
		h.Authorizations,
		h.Clock.NowUTC(),
		h.AppID,
		client.ClientID,
		info.UserID,
		scopes,
	)
	if err != nil {
		return nil, err
	}

	// Clean up any offline grants that were issued with the same identity.
	offlineGrants, err := h.OfflineGrants.ListOfflineGrants(authz.UserID)
	if err != nil {
		return nil, err
	}
	for _, offlineGrant := range offlineGrants {
		if offlineGrant.IdentityID == biometricIdentity.ID {
			err := h.OfflineGrants.DeleteOfflineGrant(offlineGrant)
			if err != nil {
				return nil, err
			}
		}
	}

	resp := protocol.TokenResponse{}

	// SSOEnabled is false for refresh tokens that are granted by biometric login
	opts := IssueOfflineGrantOptions{
		Scopes:             scopes,
		AuthorizationID:    authz.ID,
		AuthenticationInfo: info,
		DeviceInfo:         deviceInfo,
		IdentityID:         biometricIdentity.ID,
		SSOEnabled:         false,
	}
	offlineGrant, err := h.TokenService.IssueOfflineGrant(client, opts, resp)
	if err != nil {
		return nil, err
	}

	err = h.TokenService.IssueAccessGrant(client, scopes, authz.ID, authz.UserID,
		offlineGrant.ID, oauth.GrantSessionKindOffline, resp)
	if err != nil {
		return nil, err
	}

	if h.IDTokenIssuer == nil {
		return nil, errors.New("id token issuer is not provided")
	}
	idToken, err := h.IDTokenIssuer.IssueIDToken(oidc.IssueIDTokenOptions{
		ClientID:           client.ClientID,
		SID:                oidc.EncodeSID(offlineGrant),
		AuthenticationInfo: offlineGrant.GetAuthenticationInfo(),
		ClientLike: &oauth.ClientLike{
			ClientParty: client.ClientParty(),
			Scopes:      scopes,
		},
	})
	if err != nil {
		return nil, err
	}
	resp.IDToken(idToken)

	return tokenResultOK{Response: resp}, nil
}

func (h *TokenHandler) handleIDToken(
	w http.ResponseWriter,
	req *http.Request,
	client *config.OAuthClientConfig,
	r protocol.TokenRequest,
) (httputil.Result, error) {
	if client.ClientParty() == config.ClientPartyThird {
		return nil, protocol.NewError(
			"unauthorized_client",
			"third-party clients may not refresh id token",
		)
	}

	s := session.GetSession(req.Context())
	if s == nil {
		return nil, protocol.NewErrorStatusCode("invalid_request", "valid session is required", http.StatusUnauthorized)
	}
	idToken, err := h.IDTokenIssuer.IssueIDToken(oidc.IssueIDTokenOptions{
		ClientID:           client.ClientID,
		SID:                oidc.EncodeSID(s),
		AuthenticationInfo: s.GetAuthenticationInfo(),
		ClientLike: &oauth.ClientLike{
			ClientParty: client.ClientParty(),
			// scopes are used for specifying which fields should be included in the ID token
			// those fields may include personal identifiable information
			// Since the ID token issued here will be used in id_token_hint
			// so no scopes are needed
			Scopes: []string{},
		},
	})
	if err != nil {
		return nil, err
	}
	resp := protocol.TokenResponse{}
	resp.IDToken(idToken)
	return tokenResultOK{Response: resp}, nil
}

func (h *TokenHandler) issueTokensForAuthorizationCode(
	client *config.OAuthClientConfig,
	code *oauth.CodeGrant,
	authz *oauth.Authorization,
	deviceInfo map[string]interface{},
) (protocol.TokenResponse, error) {
	issueRefreshToken := false
	issueIDToken := false
	for _, scope := range code.Scopes {
		switch scope {
		case "offline_access":
			issueRefreshToken = true
		case "openid":
			issueIDToken = true
		}
	}

	if issueRefreshToken {
		// Only if client is allowed to use refresh tokens
		allowRefreshToken := false
		for _, grantType := range client.GrantTypes {
			if grantType == "refresh_token" {
				allowRefreshToken = true
				break
			}
		}
		if !allowRefreshToken {
			issueRefreshToken = false
		}
	}

	info := code.AuthenticationInfo

	// Update auth_time of the offline grant if possible.
	if sid := code.IDTokenHintSID; sid != "" {
		if typ, sessionID, ok := oidc.DecodeSID(sid); ok && typ == session.TypeOfflineGrant {
			offlineGrant, err := h.OfflineGrants.GetOfflineGrant(sessionID)
			if err == nil {
				if info.AuthenticatedAt.After(offlineGrant.AuthenticatedAt) {
					expiry, err := h.OfflineGrantService.ComputeOfflineGrantExpiry(offlineGrant)
					if err != nil {
						return nil, err
					}
					_, err = h.OfflineGrants.UpdateOfflineGrantAuthenticatedAt(offlineGrant.ID, info.AuthenticatedAt, expiry)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	resp := protocol.TokenResponse{}

	// As required by the spec, we must include access_token.
	// If we issue refresh token, then access token is just the access token of the refresh token.
	// Else if id_token_hint is present, use the sid.
	// Otherwise we return an error.
	var accessTokenSessionID string
	var accessTokenSessionKind oauth.GrantSessionKind
	var sid string

	opts := IssueOfflineGrantOptions{
		Scopes:             code.Scopes,
		AuthorizationID:    authz.ID,
		AuthenticationInfo: info,
		IDPSessionID:       code.IDPSessionID,
		DeviceInfo:         deviceInfo,
		SSOEnabled:         code.SSOEnabled,
	}
	if issueRefreshToken {
		offlineGrant, err := h.TokenService.IssueOfflineGrant(client, opts, resp)
		if err != nil {
			return nil, err
		}
		sid = oidc.EncodeSID(offlineGrant)
		accessTokenSessionID = offlineGrant.ID
		accessTokenSessionKind = oauth.GrantSessionKindOffline
	} else if code.IDTokenHintSID != "" {
		sid = code.IDTokenHintSID
		if typ, sessionID, ok := oidc.DecodeSID(sid); ok {
			accessTokenSessionID = sessionID
			switch typ {
			case session.TypeOfflineGrant:
				accessTokenSessionKind = oauth.GrantSessionKindOffline
			case session.TypeIdentityProvider:
				accessTokenSessionKind = oauth.GrantSessionKindSession
			default:
				panic(fmt.Errorf("unknown session type: %v", typ))
			}
		}
	} else if client.ClientParty() == config.ClientPartyThird {
		// allow issuing access tokens if scopes don't contain offline_access and the client is third-party
		// fill the response with nil for not returning the refresh token
		offlineGrant, err := h.TokenService.IssueOfflineGrant(client, opts, nil)
		if err != nil {
			return nil, err
		}
		sid = oidc.EncodeSID(offlineGrant)
		accessTokenSessionID = offlineGrant.ID
		accessTokenSessionKind = oauth.GrantSessionKindOffline
	}

	if accessTokenSessionID == "" || accessTokenSessionKind == "" {
		return nil, protocol.NewError("invalid_request", "cannot issue access token")
	}

	err := h.TokenService.IssueAccessGrant(client, code.Scopes, authz.ID, authz.UserID, accessTokenSessionID, accessTokenSessionKind, resp)
	if err != nil {
		return nil, err
	}

	if issueIDToken {
		if h.IDTokenIssuer == nil {
			return nil, errors.New("id token issuer is not provided")
		}
		if sid == "" {
			return nil, protocol.NewError("invalid_request", "cannot issue ID token")
		}
		idToken, err := h.IDTokenIssuer.IssueIDToken(oidc.IssueIDTokenOptions{
			ClientID:           client.ClientID,
			SID:                sid,
			Nonce:              code.OIDCNonce,
			AuthenticationInfo: info,
			ClientLike: &oauth.ClientLike{
				ClientParty: client.ClientParty(),
				Scopes:      code.Scopes,
			},
		})
		if err != nil {
			return nil, err
		}
		resp.IDToken(idToken)
	}

	return resp, nil
}

func (h *TokenHandler) issueTokensForRefreshToken(
	client *config.OAuthClientConfig,
	offlineGrant *oauth.OfflineGrant,
	authz *oauth.Authorization,
) (protocol.TokenResponse, error) {
	issueIDToken := false
	for _, scope := range offlineGrant.Scopes {
		if scope == "openid" {
			issueIDToken = true
			break
		}
	}

	resp := protocol.TokenResponse{}

	if issueIDToken {
		if h.IDTokenIssuer == nil {
			return nil, errors.New("id token issuer is not provided")
		}
		idToken, err := h.IDTokenIssuer.IssueIDToken(oidc.IssueIDTokenOptions{
			ClientID:           client.ClientID,
			SID:                oidc.EncodeSID(offlineGrant),
			AuthenticationInfo: offlineGrant.GetAuthenticationInfo(),
			ClientLike: &oauth.ClientLike{
				ClientParty: client.ClientParty(),
				Scopes:      authz.Scopes,
			},
		})
		if err != nil {
			return nil, err
		}
		resp.IDToken(idToken)
	}

	err := h.TokenService.IssueAccessGrant(client, offlineGrant.Scopes,
		authz.ID, authz.UserID, offlineGrant.ID, oauth.GrantSessionKindOffline, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type IssueOfflineGrantOptions struct {
	AuthenticationInfo authenticationinfo.T
	Scopes             []string
	AuthorizationID    string
	IDPSessionID       string
	DeviceInfo         map[string]interface{}
	IdentityID         string
	SSOEnabled         bool
}

func (h *TokenHandler) IssueAppSessionToken(refreshToken string) (string, *oauth.AppSessionToken, error) {
	authz, grant, err := h.TokenService.ParseRefreshToken(refreshToken)
	if err != nil {
		return "", nil, err
	}

	// Ensure client is authorized with full user access (i.e. first-party client)
	if !authz.IsAuthorized([]string{oauth.FullAccessScope}) {
		return "", nil, protocol.NewError("access_denied", "the client is not authorized to have full user access")
	}

	now := h.Clock.NowUTC()
	token := oauth.GenerateToken()
	sToken := &oauth.AppSessionToken{
		AppID:          grant.AppID,
		OfflineGrantID: grant.ID,
		CreatedAt:      now,
		ExpireAt:       now.Add(AppSessionTokenDuration),
		TokenHash:      oauth.HashToken(token),
	}

	err = h.AppSessionTokens.CreateAppSessionToken(sToken)
	if err != nil {
		return "", nil, err
	}

	return token, sToken, err
}

func verifyPKCE(challenge, verifier string) bool {
	verifierHash := sha256.Sum256([]byte(verifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(verifierHash[:])
	return subtle.ConstantTimeCompare([]byte(challenge), []byte(expectedChallenge)) == 1
}
