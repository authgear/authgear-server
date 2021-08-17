package handler

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	identitybiometric "github.com/authgear/authgear-server/pkg/lib/authn/identity/biometric"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	interactionintents "github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/uuid"
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
	IssueIDToken(client *config.OAuthClientConfig, session session.Session, nonce string) (token string, err error)
}

type AccessTokenIssuer interface {
	EncodeAccessToken(client *config.OAuthClientConfig, grant *oauth.AccessGrant, userID string, token string) (string, error)
}

type SessionProvider interface {
	Get(id string) (*idpsession.IDPSession, error)
}

type TokenHandlerUserFacade interface {
	GetRaw(id string) (*user.User, error)
}

type TokenHandlerLogger struct{ *log.Logger }

func NewTokenHandlerLogger(lf *log.Factory) TokenHandlerLogger {
	return TokenHandlerLogger{lf.New("oauth-token")}
}

type TokenHandler struct {
	Request    *http.Request
	AppID      config.AppID
	Config     *config.OAuthConfig
	TrustProxy config.TrustProxy
	Logger     TokenHandlerLogger

	Authorizations    oauth.AuthorizationStore
	CodeGrants        oauth.CodeGrantStore
	OfflineGrants     oauth.OfflineGrantStore
	AccessGrants      oauth.AccessGrantStore
	AppSessionTokens  oauth.AppSessionTokenStore
	AccessEvents      *access.EventProvider
	Sessions          SessionProvider
	Graphs            GraphService
	IDTokenIssuer     IDTokenIssuer
	AccessTokenIssuer AccessTokenIssuer
	GenerateToken     TokenGenerator
	Clock             clock.Clock
	Users             TokenHandlerUserFacade
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
	if err := h.validateRequest(r); err != nil {
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

func (h *TokenHandler) validateRequest(r protocol.TokenRequest) error {
	switch r.GrantType() {
	case "authorization_code":
		if r.Code() == "" {
			return protocol.NewError("invalid_request", "code is required")
		}
		if r.CodeVerifier() == "" {
			return protocol.NewError("invalid_request", "PKCE code verifier is required")
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

	if codeGrant.PKCEChallenge != "" && !verifyPKCE(codeGrant.PKCEChallenge, r.CodeVerifier()) {
		return nil, errInvalidAuthzCode
	}

	authz, err := h.Authorizations.GetByID(codeGrant.AuthorizationID)
	if errors.Is(err, oauth.ErrAuthorizationNotFound) {
		return nil, errInvalidAuthzCode
	} else if err != nil {
		return nil, err
	}

	sess, err := h.Sessions.Get(codeGrant.IDPSessionID)
	if errors.Is(err, idpsession.ErrSessionNotFound) {
		return nil, errInvalidAuthzCode
	} else if err != nil {
		return nil, err
	}

	resp, err := h.issueTokensForAuthorizationCode(client, codeGrant, authz, sess, deviceInfo)
	if err != nil {
		return nil, err
	}

	err = h.CodeGrants.DeleteCodeGrant(codeGrant)
	if err != nil {
		h.Logger.WithError(err).Error("failed to invalidate code grant")
	}

	return tokenResultOK{Response: resp}, nil
}

var errInvalidRefreshToken = protocol.NewError("invalid_grant", "invalid refresh token")

func (h *TokenHandler) parseRefreshToken(token string) (*oauth.Authorization, *oauth.OfflineGrant, error) {
	token, grantID, err := oauth.DecodeRefreshToken(token)
	if err != nil {
		return nil, nil, errInvalidRefreshToken
	}

	offlineGrant, err := h.OfflineGrants.GetOfflineGrant(grantID)
	if errors.Is(err, oauth.ErrGrantNotFound) {
		return nil, nil, errInvalidRefreshToken
	} else if err != nil {
		return nil, nil, err
	}

	expiry, err := oauth.ComputeOfflineGrantExpiryWithClients(offlineGrant, h.Config)
	if errors.Is(err, oauth.ErrGrantNotFound) {
		return nil, nil, errInvalidRefreshToken
	} else if err != nil {
		return nil, nil, err
	}

	if h.Clock.NowUTC().After(expiry) {
		return nil, nil, errInvalidRefreshToken
	}

	tokenHash := oauth.HashToken(token)
	if subtle.ConstantTimeCompare([]byte(tokenHash), []byte(offlineGrant.TokenHash)) != 1 {
		return nil, nil, errInvalidRefreshToken
	}

	authz, err := h.Authorizations.GetByID(offlineGrant.AuthorizationID)
	if errors.Is(err, oauth.ErrAuthorizationNotFound) {
		return nil, nil, errInvalidRefreshToken
	} else if err != nil {
		return nil, nil, err
	}

	// Check if the user has been disabled.
	u, err := h.Users.GetRaw(offlineGrant.Attrs.UserID)
	if err != nil {
		return nil, nil, err
	}

	err = u.CheckStatus()
	if err != nil {
		return nil, nil, errInvalidRefreshToken
	}

	return authz, offlineGrant, nil
}

func (h *TokenHandler) handleRefreshToken(
	client *config.OAuthClientConfig,
	r protocol.TokenRequest,
) (protocol.TokenResponse, error) {
	deviceInfo, err := r.DeviceInfo()
	if err != nil {
		return nil, protocol.NewError("invalid_request", err.Error())
	}

	authz, offlineGrant, err := h.parseRefreshToken(r.RefreshToken())
	if err != nil {
		return nil, err
	}

	resp, err := h.issueTokensForRefreshToken(client, offlineGrant, authz)
	if err != nil {
		return nil, err
	}

	expiry := oauth.ComputeOfflineGrantExpiryWithClient(offlineGrant, client)
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

func (h *TokenHandler) handleAnonymousRequest(
	client *config.OAuthClientConfig,
	r protocol.TokenRequest,
) (httputil.Result, error) {
	if !*client.IsFirstParty {
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
		graph, err = h.Graphs.NewGraph(ctx, interactionintents.NewIntentLogin(true))
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

	attrs := session.NewAnonymousAttrs(graph.MustGetUserID())

	err = h.Graphs.Run("", graph)
	if apierrors.IsAPIError(err) {
		return nil, protocol.NewError("invalid_request", err.Error())
	} else if err != nil {
		return nil, err
	}

	// TODO(oauth): allow specifying scopes
	scopes := []string{"openid", oauth.FullAccessScope}

	authz, err := checkAuthorization(
		h.Authorizations,
		h.Clock.NowUTC(),
		h.AppID,
		client.ClientID,
		attrs.UserID,
		scopes,
	)
	if err != nil {
		return nil, err
	}

	resp := protocol.TokenResponse{}

	opts := IssueOfflineGrantOptions{
		Scopes:          scopes,
		AuthorizationID: authz.ID,
		SessionAttrs:    attrs,
		DeviceInfo:      deviceInfo,
	}
	offlineGrant, err := h.issueOfflineGrant(client, opts, resp)
	if err != nil {
		return nil, err
	}

	err = h.issueAccessGrant(client, scopes, authz.ID, authz.UserID,
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
	if !*client.IsFirstParty {
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
		graph, err = h.Graphs.NewGraph(ctx, interactionintents.NewIntentLogin(true))
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

	attrs := session.NewBiometricAttrs(graph.MustGetUserID(), graph.GetAMR())
	biometricIdentity := graph.MustGetUserLastIdentity()

	err = h.Graphs.Run("", graph)
	if apierrors.IsAPIError(err) {
		return nil, protocol.NewError("invalid_request", err.Error())
	} else if err != nil {
		return nil, err
	}

	// TODO(oauth): allow specifying scopes
	scopes := []string{"openid", oauth.FullAccessScope}

	authz, err := checkAuthorization(
		h.Authorizations,
		h.Clock.NowUTC(),
		h.AppID,
		client.ClientID,
		attrs.UserID,
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

	opts := IssueOfflineGrantOptions{
		Scopes:          scopes,
		AuthorizationID: authz.ID,
		SessionAttrs:    attrs,
		DeviceInfo:      deviceInfo,
		IdentityID:      biometricIdentity.ID,
	}
	offlineGrant, err := h.issueOfflineGrant(client, opts, resp)
	if err != nil {
		return nil, err
	}

	err = h.issueAccessGrant(client, scopes, authz.ID, authz.UserID,
		offlineGrant.ID, oauth.GrantSessionKindOffline, resp)
	if err != nil {
		return nil, err
	}

	if h.IDTokenIssuer == nil {
		return nil, errors.New("id token issuer is not provided")
	}
	idToken, err := h.IDTokenIssuer.IssueIDToken(client, offlineGrant, "")
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
	s := session.GetSession(req.Context())
	if s == nil {
		return nil, protocol.NewErrorStatusCode("invalid_request", "valid session is required", http.StatusUnauthorized)
	}
	nonce := ""
	idToken, err := h.IDTokenIssuer.IssueIDToken(client, s, nonce)
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
	s *idpsession.IDPSession,
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

	// Update auth_time of the offline grant if possible.
	if sid := code.IDTokenHintSID; sid != "" {
		if typ, sessionID, ok := oidc.DecodeSID(sid); ok && typ == session.TypeOfflineGrant {

			offlineGrant, err := h.OfflineGrants.GetOfflineGrant(sessionID)
			if err == nil {
				if s.AuthenticatedAt.After(offlineGrant.AuthenticatedAt) {
					expiry := oauth.ComputeOfflineGrantExpiryWithClient(offlineGrant, client)
					_, err := h.OfflineGrants.UpdateOfflineGrantAuthenticatedAt(offlineGrant.ID, s.AuthenticatedAt, expiry)
					if err != nil {
						return nil, err
					}
				}
			}

		}
	}

	resp := protocol.TokenResponse{}

	// The ID token has the claim `sid`.
	// The `sid` is important for reauthentication.
	// If `sid` refers to a offline grant,
	// then the auth_time of the offline grant can be updated correctly.
	var sessionToBeUsedInIDToken session.Session

	var sessionID string
	var sessionKind oauth.GrantSessionKind
	if issueRefreshToken {
		opts := IssueOfflineGrantOptions{
			Scopes:          code.Scopes,
			AuthorizationID: authz.ID,
			SessionAttrs:    &s.Attrs,
			IDPSessionID:    s.ID,
			AuthenticatedAt: &s.AuthenticatedAt,
			DeviceInfo:      deviceInfo,
		}
		offlineGrant, err := h.issueOfflineGrant(client, opts, resp)
		if err != nil {
			return nil, err
		}
		sessionToBeUsedInIDToken = offlineGrant
		sessionID = offlineGrant.ID
		sessionKind = oauth.GrantSessionKindOffline
	} else {
		sessionToBeUsedInIDToken = s
		sessionID = s.ID
		sessionKind = oauth.GrantSessionKindSession

	}

	err := h.issueAccessGrant(client, code.Scopes,
		authz.ID, authz.UserID, sessionID, sessionKind, resp)
	if err != nil {
		return nil, err
	}

	if issueIDToken {
		if h.IDTokenIssuer == nil {
			return nil, errors.New("id token issuer is not provided")
		}
		idToken, err := h.IDTokenIssuer.IssueIDToken(client, sessionToBeUsedInIDToken, code.OIDCNonce)
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
		idToken, err := h.IDTokenIssuer.IssueIDToken(client, offlineGrant, "")
		if err != nil {
			return nil, err
		}
		resp.IDToken(idToken)
	}

	err := h.issueAccessGrant(client, offlineGrant.Scopes,
		authz.ID, authz.UserID, offlineGrant.ID, oauth.GrantSessionKindOffline, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type IssueOfflineGrantOptions struct {
	Scopes          []string
	AuthorizationID string
	SessionAttrs    *session.Attrs
	IDPSessionID    string
	AuthenticatedAt *time.Time
	DeviceInfo      map[string]interface{}
	IdentityID      string
}

func (h *TokenHandler) issueOfflineGrant(
	client *config.OAuthClientConfig,
	opts IssueOfflineGrantOptions,
	resp protocol.TokenResponse,
) (*oauth.OfflineGrant, error) {
	token := h.GenerateToken()
	now := h.Clock.NowUTC()
	accessEvent := access.NewEvent(now, h.Request, bool(h.TrustProxy))

	authenticatedAt := now
	if opts.AuthenticatedAt != nil {
		authenticatedAt = *opts.AuthenticatedAt
	}

	offlineGrant := &oauth.OfflineGrant{
		AppID:           string(h.AppID),
		ID:              uuid.New(),
		Labels:          make(map[string]interface{}),
		AuthorizationID: opts.AuthorizationID,
		ClientID:        client.ClientID,
		IDPSessionID:    opts.IDPSessionID,
		IdentityID:      opts.IdentityID,

		CreatedAt:       now,
		AuthenticatedAt: authenticatedAt,
		Scopes:          opts.Scopes,
		TokenHash:       oauth.HashToken(token),

		Attrs: *opts.SessionAttrs,
		AccessInfo: access.Info{
			InitialAccess: accessEvent,
			LastAccess:    accessEvent,
		},

		DeviceInfo: opts.DeviceInfo,
	}

	expiry := oauth.ComputeOfflineGrantExpiryWithClient(offlineGrant, client)
	err := h.OfflineGrants.CreateOfflineGrant(offlineGrant, expiry)
	if err != nil {
		return nil, err
	}

	err = h.AccessEvents.InitStream(offlineGrant.ID, &offlineGrant.AccessInfo.InitialAccess)
	if err != nil {
		return nil, err
	}

	resp.RefreshToken(oauth.EncodeRefreshToken(token, offlineGrant.ID))
	return offlineGrant, nil
}

func (h *TokenHandler) issueAccessGrant(
	client *config.OAuthClientConfig,
	scopes []string,
	authzID string,
	userID string,
	sessionID string,
	sessionKind oauth.GrantSessionKind,
	resp protocol.TokenResponse,
) error {
	token := h.GenerateToken()
	now := h.Clock.NowUTC()

	accessGrant := &oauth.AccessGrant{
		AppID:           string(h.AppID),
		AuthorizationID: authzID,
		SessionID:       sessionID,
		SessionKind:     sessionKind,
		CreatedAt:       now,
		ExpireAt:        now.Add(client.AccessTokenLifetime.Duration()),
		Scopes:          scopes,
		TokenHash:       oauth.HashToken(token),
	}
	err := h.AccessGrants.CreateAccessGrant(accessGrant)
	if err != nil {
		return err
	}

	at, err := h.AccessTokenIssuer.EncodeAccessToken(client, accessGrant, userID, token)
	if err != nil {
		return err
	}

	resp.TokenType("Bearer")
	resp.AccessToken(at)
	resp.ExpiresIn(int(client.AccessTokenLifetime))
	return nil
}

func (h *TokenHandler) IssueAppSessionToken(refreshToken string) (string, *oauth.AppSessionToken, error) {
	authz, grant, err := h.parseRefreshToken(refreshToken)
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
