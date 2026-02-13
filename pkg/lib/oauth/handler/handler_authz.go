package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/settingsaction"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/otelutil"
	"github.com/authgear/authgear-server/pkg/util/pkce"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

//go:generate go tool mockgen -source=handler_authz.go -destination=handler_authz_mock_test.go -package handler_test

const (
	CodeResponseTypeElement          = "code"
	NoneResponseTypeElement          = "none"
	TokenResponseTypeElement         = "token"
	SettingsActonResponseTypeElement = "urn:authgear:params:oauth:response-type:settings-action"
	// nolint:gosec
	PreAuthenticatedURLResponseTypeElement = "urn:authgear:params:oauth:response-type:pre-authenticated-url"
)

var (
	CodeResponseType                     = protocol.NewResponseType([]string{CodeResponseTypeElement})
	NoneResponseType                     = protocol.NewResponseType([]string{NoneResponseTypeElement})
	TokenResponseType                    = protocol.NewResponseType([]string{TokenResponseTypeElement})
	SettingsActonResponseType            = protocol.NewResponseType([]string{SettingsActonResponseTypeElement})
	PreAuthenticatedURLTokenResponseType = protocol.NewResponseType([]string{PreAuthenticatedURLResponseTypeElement, TokenResponseTypeElement})
)

// whitelistedResponseTypes is a list of response types that would be always allowed
// to all clients.
var whitelistedResponseTypes = []protocol.ResponseType{
	CodeResponseType,
	NoneResponseType,
	TokenResponseType,
	SettingsActonResponseType,
	PreAuthenticatedURLTokenResponseType,
}

const CodeGrantValidDuration = duration.Short
const SettingsActionGrantValidDuration = duration.Short

type UIInfoResolver interface {
	ResolveForAuthorizationEndpoint(ctx context.Context, client *config.OAuthClientConfig, req protocol.AuthorizationRequest) (*oidc.UIInfo, *oidc.UIInfoByProduct, error)
}

type AuthenticationInfoResolver interface {
	GetAuthenticationInfoID(req *http.Request) (string, bool)
}

type UIURLBuilder interface {
	BuildAuthenticationURL(client *config.OAuthClientConfig, r protocol.AuthorizationRequest, e *oauthsession.Entry) (*url.URL, error)
	BuildSettingsActionURL(client *config.OAuthClientConfig, r protocol.AuthorizationRequest, e *oauthsession.Entry) (*url.URL, error)
}

type AppSessionTokenService interface {
	Handle(ctx context.Context, input oauth.AppSessionTokenInput) (httputil.Result, error)
}

type AuthenticationInfoService interface {
	Get(ctx context.Context, entryID string) (*authenticationinfo.Entry, error)
	Delete(ctx context.Context, entryID string) error
}

type CookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type OAuthSessionService interface {
	Save(ctx context.Context, entry *oauthsession.Entry) (err error)
	Get(ctx context.Context, entryID string) (*oauthsession.Entry, error)
	Delete(ctx context.Context, entryID string) error
}

type AuthorizationService interface {
	GetByID(ctx context.Context, id string) (*oauth.Authorization, error)
	CheckAndGrant(
		ctx context.Context,
		clientID string,
		userID string,
		scopes []string,
	) (*oauth.Authorization, error)
	Check(
		ctx context.Context,
		clientID string,
		userID string,
		scopes []string,
	) (*oauth.Authorization, error)
}

type AuthorizationHandlerAccessTokenEncoding interface {
	MakeUserAccessTokenFromPreparationResult(
		ctx context.Context,
		options oauth.MakeUserAccessTokenFromPreparationOptions,
	) (*oauth.IssueAccessGrantResult, error)
}

type AuthorizationHandlerPreAuthenticatedURLTokenService interface {
	ExchangeForAccessToken(
		ctx context.Context,
		client *config.OAuthClientConfig,
		sessionID string,
		token string,
	) (oauth.PrepareUserAccessTokenResult, error)
}

type AuthorizationHandlerDatabase interface {
	WithTx(ctx context.Context, do func(ctx context.Context) error) (err error)
}

var AuthorizationHandlerLogger = slogutil.NewLogger("oauth-authz")

type AuthorizationHandler struct {
	AppID                 config.AppID
	Config                *config.OAuthConfig
	AccountDeletionConfig *config.AccountDeletionConfig
	HTTPConfig            *config.HTTPConfig
	HTTPProto             httputil.HTTPProto
	HTTPOrigin            httputil.HTTPOrigin
	AppDomains            config.AppDomains

	Database AuthorizationHandlerDatabase

	UIURLBuilder                            UIURLBuilder
	UIInfoResolver                          UIInfoResolver
	AuthenticationInfoResolver              AuthenticationInfoResolver
	Authorizations                          AuthorizationService
	AppSessionTokenService                  AppSessionTokenService
	AuthenticationInfoService               AuthenticationInfoService
	Clock                                   clock.Clock
	Cookies                                 CookieManager
	OAuthSessionService                     OAuthSessionService
	CodeGrantService                        CodeGrantService
	SettingsActionGrantService              SettingsActionGrantService
	ClientResolver                          OAuthClientResolver
	PreAuthenticatedURLTokenService         AuthorizationHandlerPreAuthenticatedURLTokenService
	IDTokenIssuer                           IDTokenIssuer
	AuthorizationHandlerAccessTokenEncoding AuthorizationHandlerAccessTokenEncoding
}

func (h *AuthorizationHandler) HandleConsentWithoutUserConsent(ctx context.Context, req *http.Request) (httputil.Result, *ConsentRequired) {
	result, consentRequired := h.doHandleConsent(ctx, req, false)
	return result, consentRequired
}

func (h *AuthorizationHandler) HandleConsentWithUserConsent(ctx context.Context, req *http.Request) httputil.Result {
	result, _ := h.doHandleConsent(ctx, req, true)
	return result
}

func (h *AuthorizationHandler) HandleConsentWithUserCancel(ctx context.Context, req *http.Request) httputil.Result {
	logger := AuthorizationHandlerLogger.GetLogger(ctx)
	ctx, consentRequest, err := h.prepareConsentRequest(ctx, req)
	if err != nil {
		var oauthError *protocol.OAuthProtocolError
		var resultErr httputil.Result

		if errors.As(err, &oauthError) {
			resultErr = h.prepareConsentErrInvalidOAuthResponse(ctx, req, *oauthError)
		} else {
			logger.WithError(err).Error(ctx, "authz handler failed")
			resultErr = AuthorizationResultError{
				// Don't redirect for those unexpected errors
				// e.g. oauth session expire or invalid client_id, redirect_uri
				RedirectURI:   nil,
				Response:      protocol.NewErrorResponse("server_error", "internal server error"),
				InternalError: true,
			}
		}

		return resultErr
	}

	oauthSessionEntry := consentRequest.OAuthSessionEntry
	authInfoEntry := consentRequest.AuthInfoEntry
	authzReq := oauthSessionEntry.T.AuthorizationRequest
	redirectURI := consentRequest.RedirectURI
	// delete oauth session and auth info with best effort
	// don't block the user in case of failure
	err = h.OAuthSessionService.Delete(ctx, oauthSessionEntry.ID)
	if err != nil {
		logger.WithError(err).Error(ctx, "failed to consume oauth session")
	}
	err = h.AuthenticationInfoService.Delete(ctx, authInfoEntry.ID)
	if err != nil {
		logger.WithError(err).Error(ctx, "failed to consume authentication info")
	}

	resultErr := AuthorizationResultError{
		ResponseMode: authzReq.ResponseMode(),
		RedirectURI:  redirectURI,
		Response:     protocol.NewErrorResponse("access_denied", "authorization denied"),
	}
	state := authzReq.State()
	if state != "" {
		resultErr.Response.State(authzReq.State())
	}
	return resultErr
}

type ConsentRequired struct {
	UserID string
	Scopes []string
	Client *config.OAuthClientConfig
}

func (h *AuthorizationHandler) doHandleConsent(ctx context.Context, req *http.Request, withUserConsent bool) (httputil.Result, *ConsentRequired) {
	logger := AuthorizationHandlerLogger.GetLogger(ctx)
	ctx, consentRequest, err := h.prepareConsentRequest(ctx, req)
	if err != nil {
		var oauthError *protocol.OAuthProtocolError
		var resultErr httputil.Result

		if errors.As(err, &oauthError) {
			resultErr = h.prepareConsentErrInvalidOAuthResponse(ctx, req, *oauthError)
		} else {
			logger.WithError(err).Error(ctx, "authz handler failed")
			resultErr = AuthorizationResultError{
				// Don't redirect for those unexpected errors
				// e.g. oauth session expire or invalid client_id, redirect_uri
				RedirectURI:   nil,
				Response:      protocol.NewErrorResponse("server_error", "internal server error"),
				InternalError: true,
			}
		}

		return resultErr, nil
	}

	autoGrantAuthz := consentRequest.Client.IsFirstParty()
	grantAuthz := autoGrantAuthz || withUserConsent

	result, err := h.doHandleConsentRequest(ctx, doHandleConsentRequestOptions{
		ConsentRequest: consentRequest,
		HTTPRequest:    req,
		GrantAuthz:     grantAuthz,
	})
	if err != nil {
		if !grantAuthz && IsConsentRequiredError(err) {
			return nil, &ConsentRequired{
				UserID: consentRequest.AuthInfoEntry.T.UserID,
				Scopes: consentRequest.OAuthSessionEntry.T.AuthorizationRequest.Scope(),
				Client: consentRequest.Client,
			}
		}

		var oauthError *protocol.OAuthProtocolError
		resultErr := AuthorizationResultError{
			RedirectURI:  consentRequest.RedirectURI,
			ResponseMode: consentRequest.OAuthSessionEntry.T.AuthorizationRequest.ResponseMode(),
		}
		if errors.As(err, &oauthError) {
			resultErr.Response = oauthError.Response
		} else {
			logger.WithError(err).Error(ctx, "authz handler failed")
			resultErr.Response = protocol.NewErrorResponse("server_error", "internal server error")
			resultErr.InternalError = true
		}
		state := consentRequest.OAuthSessionEntry.T.AuthorizationRequest.State()
		if state != "" {
			resultErr.Response.State(consentRequest.OAuthSessionEntry.T.AuthorizationRequest.State())
		}
		result = resultErr
	} else {
		// delete oauth session with best effort
		// don't block the user in case of failure
		err = h.OAuthSessionService.Delete(ctx, consentRequest.OAuthSessionEntry.ID)
		if err != nil {
			logger.WithError(err).Error(ctx, "failed to consume oauth session")
		}
		err = h.AuthenticationInfoService.Delete(ctx, consentRequest.AuthInfoEntry.ID)
		if err != nil {
			logger.WithError(err).Error(ctx, "failed to consume authentication info")
		}
	}

	return result, nil
}

func (h *AuthorizationHandler) getAuthenticationInfoEntry(ctx context.Context, req *http.Request) (*authenticationinfo.Entry, error) {
	id, ok := h.AuthenticationInfoResolver.GetAuthenticationInfoID(req)
	if !ok {
		return nil, protocol.NewError("login_required", "authentication required")
	}

	entry, err := h.AuthenticationInfoService.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

type consentRequest struct {
	OAuthSessionEntry *oauthsession.Entry
	AuthInfoEntry     *authenticationinfo.Entry
	RedirectURI       *url.URL
	Client            *config.OAuthClientConfig
}

func (h *AuthorizationHandler) prepareConsentRequest(ctx context.Context, req *http.Request) (context.Context, *consentRequest, error) {
	authInfoEntry, err := h.getAuthenticationInfoEntry(ctx, req)
	if err != nil {
		if errors.Is(err, authenticationinfo.ErrNotFound) {
			err = protocol.NewError("invalid_request", "authentication expired")
		}
		return ctx, nil, err
	}

	entry, err := h.OAuthSessionService.Get(ctx, authInfoEntry.OAuthSessionID)
	if err != nil {
		if errors.Is(err, oauthsession.ErrNotFound) {
			err = protocol.NewError("invalid_request", "oauth session expired")
		}
		return ctx, nil, err
	}

	r := entry.T.AuthorizationRequest

	ctx, client := resolveClient(ctx, h.ClientResolver, r.ClientID())
	if client == nil {
		err = protocol.NewError("unauthorized_client", "invalid client ID")
		return ctx, nil, err
	}

	uiInfo, _, err := h.UIInfoResolver.ResolveForAuthorizationEndpoint(ctx, client, r)
	if err != nil {
		return ctx, nil, err
	}

	uiParam := uiInfo.ToUIParam()
	// Restore uiparam into context.
	uiparam.WithUIParam(ctx, &uiParam)

	redirectURI, errResp := parseRedirectURI(client, h.HTTPProto, h.HTTPOrigin, h.AppDomains, []string{}, r)
	if errResp != nil {
		err = protocol.NewErrorWithErrorResponse(errResp)
		return ctx, nil, err
	}

	return ctx, &consentRequest{
		OAuthSessionEntry: entry,
		AuthInfoEntry:     authInfoEntry,
		RedirectURI:       redirectURI,
		Client:            client,
	}, nil
}

func (h *AuthorizationHandler) HandleRequest(
	ctx context.Context,
	r protocol.AuthorizationRequest,
	params *AuthorizationParams,
) (result httputil.Result) {
	logger := AuthorizationHandlerLogger.GetLogger(ctx)
	var err error

	defer func() {
		if err != nil {
			var oauthError *protocol.OAuthProtocolError
			resultErr := AuthorizationResultError{
				RedirectURI:  params.RedirectURI,
				ResponseMode: r.ResponseMode(),
			}
			if errors.As(err, &oauthError) {
				resultErr.Response = oauthError.Response
			} else {
				logger.WithError(err).Error(ctx, "authz handler failed")
				resultErr.Response = protocol.NewErrorResponse("server_error", "internal server error")
				resultErr.InternalError = true
			}
			state := r.State()
			if state != "" {
				resultErr.Response.State(r.State())
			}
			result = resultErr
		}
	}()

	if r.ResponseType().Equal(PreAuthenticatedURLTokenResponseType) {
		var preparationResult oauth.PrepareUserAccessTokenResult
		err = h.Database.WithTx(ctx, func(ctx context.Context) error {
			preparationResult, err = h.doHandlePreAuthenticatedURLWithTx(ctx, params.RedirectURI, params.Client, r)
			return err
		})
		if err != nil {
			return
		}
		result, err = h.doHandlePreAuthenticatedURLAfterTx(ctx, params.RedirectURI, params.Client, r, preparationResult)
		if err != nil {
			return
		}
	} else {
		err = h.Database.WithTx(ctx, func(ctx context.Context) error {
			result, err = h.doHandleRequestWithTx(ctx, params.RedirectURI, params.Client, r)
			return err
		})
		if err != nil {
			return
		}
	}

	return
}

// nolint: gocognit
func (h *AuthorizationHandler) doHandleRequestWithTx(
	ctx context.Context,
	redirectURI *url.URL,
	client *config.OAuthClientConfig,
	r protocol.AuthorizationRequest,
) (httputil.Result, error) {
	err := oauth.ValidateScopesByClientConfig(client, r.Scope())
	if err != nil {
		return nil, err
	}

	if r.ResponseType().Equal(SettingsActonResponseType) {
		// create oauth session for the setting action
		oauthSessionEntry := oauthsession.NewEntry(oauthsession.T{
			AuthorizationRequest: r,
			SettingsActionID:     settingsaction.NewSettingsActionID(),
		})
		err = h.OAuthSessionService.Save(ctx, oauthSessionEntry)
		if err != nil {
			return nil, err
		}
		return h.handleSettingsAction(
			ctx,
			redirectURI,
			client,
			oauthSessionEntry,
			r,
		)
	}

	// create oauth session and redirect to the web app
	oauthSessionEntry := oauthsession.NewEntry(oauthsession.T{
		AuthorizationRequest: r,
	})
	err = h.OAuthSessionService.Save(ctx, oauthSessionEntry)
	if err != nil {
		return nil, err
	}

	otelutil.IntCounterAddOne(
		ctx,
		otelauthgear.CounterOAuthSessionCreationCount,
	)

	loginHintString, loginHintOk := r.LoginHint()
	// Handle app session token here, and return here.
	// Anonymous user promotion is handled by the normal flow below.
	if loginHintOk {
		loginHint, err := oauth.ParseLoginHint(loginHintString)
		if err != nil {
			return nil, protocol.NewError("invalid_request", err.Error())
		}

		if loginHint.Type == oauth.LoginHintTypeAppSessionToken {
			result, err := h.AppSessionTokenService.Handle(ctx, oauth.AppSessionTokenInput{
				AppSessionToken: loginHint.AppSessionToken,
				RedirectURI:     redirectURI.String(),
			})
			if err != nil {
				return nil, protocol.NewError("invalid_request", err.Error())
			}
			return result, nil
		}
	}

	uiInfo, uiInfoByProduct, err := h.UIInfoResolver.ResolveForAuthorizationEndpoint(ctx, client, r)
	if err != nil {
		return nil, err
	}
	idToken := uiInfoByProduct.IDToken
	idTokenHintSID := uiInfoByProduct.IDTokenHintSID

	// Handle prompt!=none
	// We must return here.
	if !slice.ContainsString(uiInfo.Prompt, "none") {
		endpoint, err := h.UIURLBuilder.BuildAuthenticationURL(client, r, oauthSessionEntry)
		if apierrors.IsKind(err, oidc.ErrInvalidCustomURI) {
			return nil, protocol.NewError("invalid_request", err.Error())
		} else if err != nil {
			return nil, err
		}

		resp := &httputil.InternalRedirectResult{
			Cookies: []*http.Cookie{
				h.Cookies.ValueCookie(oauthsession.UICookieDef, oauthSessionEntry.ID),
			},
			URL: endpoint.String(),
		}
		return resp, nil
	}

	// Handle prompt=none
	var resolvedSession session.ResolvedSession
	if s := session.GetSession(ctx); s != nil {
		resolvedSession = s
	}
	// Ignore any session that is not allow to be used here
	if !oauth.ContainsAllScopes(oauth.SessionScopes(resolvedSession), []string{oauth.PreAuthenticatedURLScope}) {
		resolvedSession = nil
	}
	if resolvedSession == nil || (idToken != nil && resolvedSession.GetAuthenticationInfo().UserID != idToken.Subject()) {
		return nil, protocol.NewError("login_required", "authentication required")
	}

	authenticationInfo := resolvedSession.CreateNewAuthenticationInfoByThisSession()
	autoGrantAuthz := client.IsFirstParty()

	sessionType := resolvedSession.SessionType()
	sessionID := resolvedSession.SessionID()

	result, err := h.finishAuthorization(ctx, FinishAuthorizationOptions{
		Client:               client,
		RedirectURI:          redirectURI,
		AuthorizationRequest: r,
		SessionType:          sessionType,
		SessionID:            sessionID,
		AuthenticationInfo:   authenticationInfo,
		IDTokenHintSID:       idTokenHintSID,
		Cookies:              nil,
		GrantAuthz:           autoGrantAuthz,
	})
	if err != nil {
		if errors.Is(err, oauth.ErrAuthorizationNotFound) {
			return nil, protocol.NewError("access_denied", "authorization required")
		}
		if errors.Is(err, oauth.ErrAuthorizationScopesNotGranted) {
			return nil, protocol.NewError("access_denied", "requested scopes are not granted")
		}
		return nil, err
	}
	return result, nil
}

func (h *AuthorizationHandler) doHandlePreAuthenticatedURLWithTx(
	ctx context.Context,
	redirectURI *url.URL,
	client *config.OAuthClientConfig,
	r protocol.AuthorizationRequest,
) (oauth.PrepareUserAccessTokenResult, error) {
	idTokenHint, ok := r.IDTokenHint()
	if !ok {
		panic("cannot get id_token_hint, the request should be validated")
	}
	idToken, err := h.IDTokenIssuer.VerifyIDToken(idTokenHint)
	if err != nil {
		return nil, protocol.NewError("invalid_request", "invalid id_token_hint")
	}
	var sidInt interface{}
	if sidInt, ok = idToken.Get(string(model.ClaimSID)); !ok {
		return nil, protocol.NewError("invalid_request", "required sid in id_token_hint")
	}
	var sid string
	if sid, ok = sidInt.(string); !ok {
		return nil, protocol.NewError("invalid_request", "sid is not a string in id_token_hint")
	}
	_, sessionID, ok := oauth.DecodeSID(sid)
	if !ok {
		return nil, protocol.NewError("invalid_request", "invalid sid format id_token_hint")
	}

	preparationResult, err := h.PreAuthenticatedURLTokenService.ExchangeForAccessToken(
		ctx,
		client,
		sessionID,
		r.PreAuthenticatedURLToken(),
	)
	if err != nil {
		if errors.Is(err, oauth.ErrUnmatchedClient) {
			return nil, protocol.NewError("invalid_request", "incorrect client_id")
		}
		if errors.Is(err, oauth.ErrUnmatchedSession) {
			return nil, protocol.NewError("invalid_request", "incorrect sid in id_token_hint")
		}
		if errors.Is(err, oauth.ErrGrantNotFound) {
			return nil, protocol.NewError("invalid_request", "invalid x_pre_authenticated_url_token")
		}
		return nil, err
	}

	return preparationResult, nil
}

func (h *AuthorizationHandler) doHandlePreAuthenticatedURLAfterTx(
	ctx context.Context,
	redirectURI *url.URL,
	client *config.OAuthClientConfig,
	r protocol.AuthorizationRequest,
	preparationResult oauth.PrepareUserAccessTokenResult,
) (httputil.Result, error) {
	accessTokenResult, err := h.AuthorizationHandlerAccessTokenEncoding.MakeUserAccessTokenFromPreparationResult(ctx, oauth.MakeUserAccessTokenFromPreparationOptions{
		PreparationResult: preparationResult,
	})
	if err != nil {
		return nil, err
	}

	cookie := h.Cookies.ValueCookie(session.AppAccessTokenCookieDef, accessTokenResult.Token)

	resp := protocol.AuthorizationResponse{}
	state := r.State()
	if state != "" {
		resp.State(r.State())
	}
	return authorizationResultCode{
		RedirectURI:  redirectURI,
		ResponseMode: r.ResponseMode(),
		UseHTTP200:   client.UseHTTP200(),
		Response:     resp,
		Cookies:      []*http.Cookie{cookie},
	}, nil
}

func (h *AuthorizationHandler) handleSettingsAction(
	ctx context.Context,
	redirectURI *url.URL,
	client *config.OAuthClientConfig,
	oauthSessionEntry *oauthsession.Entry,
	r protocol.AuthorizationRequest,
) (httputil.Result, error) {
	redirectURI, err := h.UIURLBuilder.BuildSettingsActionURL(client, r, oauthSessionEntry)
	if err != nil {
		return nil, err
	}
	loginHintString, loginHintOk := r.LoginHint()
	if !loginHintOk {
		return nil, protocol.NewError("invalid_request", "login_hint must be provided when using settings action")
	}
	loginHint, err := oauth.ParseLoginHint(loginHintString)
	if err != nil {
		return nil, protocol.NewError("invalid_request", err.Error())
	}

	if loginHint.Type != oauth.LoginHintTypeAppSessionToken {
		return nil, protocol.NewError("invalid_request", "login_hint must be app_session_token when using settings action")
	}

	result, err := h.AppSessionTokenService.Handle(ctx, oauth.AppSessionTokenInput{
		AppSessionToken: loginHint.AppSessionToken,
		RedirectURI:     redirectURI.String(),
	})
	if err != nil {
		return nil, protocol.NewError("invalid_request", err.Error())
	}
	return result, nil

}

func (h *AuthorizationHandler) finishSettingsAction(
	ctx context.Context,
	client *config.OAuthClientConfig,
	redirectURI *url.URL,
	r protocol.AuthorizationRequest,
	cookies []*http.Cookie,
	userID string,
) (httputil.Result, error) {
	resp := protocol.AuthorizationResponse{}
	responseType := r.ResponseType()
	switch {
	case responseType.Equal(SettingsActonResponseType):
		err := h.generateSettingsActionResponse(ctx, redirectURI.String(), r, resp, userID)
		if err != nil {
			return nil, err
		}
	default:
		panic("oauth: unexpected response type. This method should only be used for settings action.")
	}

	state := r.State()
	if state != "" {
		resp.State(r.State())
	}

	return authorizationResultCode{
		RedirectURI:  redirectURI,
		ResponseMode: r.ResponseMode(),
		UseHTTP200:   client.UseHTTP200(),
		Response:     resp,
		Cookies:      cookies,
	}, nil
}

type FinishAuthorizationOptions struct {
	Client               *config.OAuthClientConfig
	RedirectURI          *url.URL
	AuthorizationRequest protocol.AuthorizationRequest
	SessionType          session.Type
	SessionID            string
	AuthenticationInfo   authenticationinfo.T
	IDTokenHintSID       string
	Cookies              []*http.Cookie
	GrantAuthz           bool
}

func (h *AuthorizationHandler) finishAuthorization(
	ctx context.Context,
	opts FinishAuthorizationOptions,
) (httputil.Result, error) {
	var authz *oauth.Authorization
	var err error
	if opts.GrantAuthz {
		authz, err = h.Authorizations.CheckAndGrant(
			ctx,
			opts.AuthorizationRequest.ClientID(),
			opts.AuthenticationInfo.UserID,
			opts.AuthorizationRequest.Scope(),
		)
	} else {
		authz, err = h.Authorizations.Check(
			ctx,
			opts.AuthorizationRequest.ClientID(),
			opts.AuthenticationInfo.UserID,
			opts.AuthorizationRequest.Scope(),
		)
	}
	if err != nil {
		return nil, err
	}

	resp := protocol.AuthorizationResponse{}
	responseType := opts.AuthorizationRequest.ResponseType()
	switch {
	case responseType.Equal(CodeResponseType):
		err = h.generateCodeResponse(
			ctx,
			&CreateCodeGrantOptions{
				Authorization:        authz,
				SessionType:          opts.SessionType,
				SessionID:            opts.SessionID,
				AuthenticationInfo:   opts.AuthenticationInfo,
				IDTokenHintSID:       opts.IDTokenHintSID,
				RedirectURI:          opts.RedirectURI.String(),
				AuthorizationRequest: opts.AuthorizationRequest,
				DPoPJKT:              opts.AuthorizationRequest.DPoPJKT(),
			},
			resp,
		)
		if err != nil {
			return nil, err
		}

	case responseType.Equal(NoneResponseType):
		break

	default:
		panic("oauth: unexpected response type")
	}

	state := opts.AuthorizationRequest.State()
	if state != "" {
		resp.State(opts.AuthorizationRequest.State())
	}

	return authorizationResultCode{
		RedirectURI:  opts.RedirectURI,
		ResponseMode: opts.AuthorizationRequest.ResponseMode(),
		UseHTTP200:   opts.Client.UseHTTP200(),
		Response:     resp,
		Cookies:      opts.Cookies,
	}, nil
}

type doHandleConsentRequestOptions struct {
	ConsentRequest *consentRequest
	HTTPRequest    *http.Request
	GrantAuthz     bool
}

func (h *AuthorizationHandler) doHandleConsentRequest(
	ctx context.Context,
	opts doHandleConsentRequestOptions,
) (httputil.Result, error) {
	if err := h.doValidateRequestWithoutTx(
		opts.ConsentRequest.Client,
		opts.ConsentRequest.OAuthSessionEntry.T.AuthorizationRequest,
	); err != nil {
		return nil, err
	}

	err := oauth.ValidateScopesByClientConfig(
		opts.ConsentRequest.Client,
		opts.ConsentRequest.OAuthSessionEntry.T.AuthorizationRequest.Scope(),
	)
	if err != nil {
		return nil, err
	}

	responseType := opts.ConsentRequest.OAuthSessionEntry.T.AuthorizationRequest.ResponseType()
	switch {
	case responseType.Equal(SettingsActonResponseType):
		userID := opts.ConsentRequest.AuthInfoEntry.T.UserID
		return h.finishSettingsAction(
			ctx,
			opts.ConsentRequest.Client,
			opts.ConsentRequest.RedirectURI,
			opts.ConsentRequest.OAuthSessionEntry.T.AuthorizationRequest,
			[]*http.Cookie{},
			userID,
		)
	default:
		_, uiInfoByProduct, err := h.UIInfoResolver.ResolveForAuthorizationEndpoint(
			ctx,
			opts.ConsentRequest.Client,
			opts.ConsentRequest.OAuthSessionEntry.T.AuthorizationRequest,
		)
		if err != nil {
			return nil, err
		}
		idTokenHintSID := uiInfoByProduct.IDTokenHintSID

		sessionID := ""
		var sessionType session.Type = ""

		if opts.ConsentRequest.AuthInfoEntry.T.AuthenticatedBySessionID != "" {
			sessionID = opts.ConsentRequest.AuthInfoEntry.T.AuthenticatedBySessionID
			sessionType = session.Type(opts.ConsentRequest.AuthInfoEntry.T.AuthenticatedBySessionType)
		}

		return h.finishAuthorization(ctx, FinishAuthorizationOptions{
			Client:               opts.ConsentRequest.Client,
			RedirectURI:          opts.ConsentRequest.RedirectURI,
			AuthorizationRequest: opts.ConsentRequest.OAuthSessionEntry.T.AuthorizationRequest,
			SessionType:          sessionType,
			SessionID:            sessionID,
			AuthenticationInfo:   opts.ConsentRequest.AuthInfoEntry.T,
			IDTokenHintSID:       idTokenHintSID,
			Cookies:              []*http.Cookie{},
			GrantAuthz:           opts.GrantAuthz,
		})
	}
}

func (h *AuthorizationHandler) validatePreAuthenticatedURLTokenRequest(
	client *config.OAuthClientConfig,
	r protocol.AuthorizationRequest,
) error {
	if len(r.Prompt()) != 1 || r.Prompt()[0] != "none" {
		return protocol.NewError("invalid_request", "only 'prompt=none' is supported when using pre-authenticated url")
	}
	if idTokenHint, ok := r.IDTokenHint(); !ok || idTokenHint == "" {
		return protocol.NewError("invalid_request", "id_token_hint is required when using pre-authenticated url")
	}
	if r.PreAuthenticatedURLToken() == "" {
		return protocol.NewError("invalid_request", "x_pre_authenticated_url_token is required when using pre-authenticated url")
	}
	if r.ResponseMode() != "cookie" {
		return protocol.NewError("invalid_request", "only 'response_mode=cookie' is supported when using pre-authenticated url")
	}
	return nil
}

// nolint:gocognit
func (h *AuthorizationHandler) ValidateRequestWithoutTx(
	ctx context.Context,
	r protocol.AuthorizationRequest,
) (context.Context, *AuthorizationParams, *AuthorizationResultError) {
	ctx, client := resolveClient(ctx, h.ClientResolver, r.ClientID())
	if client == nil {
		return ctx, nil, &AuthorizationResultError{
			ResponseMode: r.ResponseMode(),
			Response:     protocol.NewErrorResponse("unauthorized_client", "invalid client ID"),
		}
	}

	switch client.ApplicationType {
	case config.OAuthClientApplicationTypeM2M:
		return ctx, nil, &AuthorizationResultError{
			ResponseMode: r.ResponseMode(),
			Response:     protocol.NewErrorResponse("unauthorized_client", "m2m clients are not allowed to use the authorize endpoint"),
		}
	default:
		originWhitelist := []string{}
		if r.ResponseType().Equal(PreAuthenticatedURLTokenResponseType) {
			originWhitelist = client.PreAuthenticatedURLAllowedOrigins
		}

		redirectURI, errResp := parseRedirectURI(client, h.HTTPProto, h.HTTPOrigin, h.AppDomains, originWhitelist, r)
		if errResp != nil {
			return ctx, nil, &AuthorizationResultError{
				ResponseMode: r.ResponseMode(),
				Response:     errResp,
			}
		}

		if err := h.doValidateRequestWithoutTx(client, r); err != nil {
			var oauthError *protocol.OAuthProtocolError
			resultErr := AuthorizationResultError{
				RedirectURI:  redirectURI,
				ResponseMode: r.ResponseMode(),
			}
			if errors.As(err, &oauthError) {
				resultErr.Response = oauthError.Response
			} else {
				resultErr.Response = protocol.NewErrorResponse("server_error", "internal server error")
				resultErr.InternalError = true
			}
			state := r.State()
			if state != "" {
				resultErr.Response.State(r.State())
			}
			return ctx, nil, &resultErr
		}

		return ctx, &AuthorizationParams{
			Client:      client,
			RedirectURI: redirectURI,
		}, nil
	}
}

func (h *AuthorizationHandler) validateResponseTypeIsWhitelisted(r protocol.AuthorizationRequest) error {
	ok := false
	responseType := r.ResponseType()
	for _, respType := range whitelistedResponseTypes {
		if respType.Equal(responseType) {
			ok = true
			break
		}
	}
	if !ok {
		return protocol.NewError("unauthorized_client", "response type is not allowed for this client")
	}
	return nil
}

func (h *AuthorizationHandler) validatePrompt(r protocol.AuthorizationRequest) error {
	if slice.ContainsString(r.Prompt(), "none") {
		if len(r.Prompt()) != 1 {
			return protocol.NewError("invalid_request", "prompt cannot have other values when none is set")
		}
		if r.HasMaxAge() {
			return protocol.NewError("invalid_request", "max_age could imply prompt=login so max_age cannot be present when prompt=none")
		}
	}
	return nil
}

func (h *AuthorizationHandler) validateRequestParameters(
	client *config.OAuthClientConfig,
	r protocol.AuthorizationRequest,
) error {
	responseType := r.ResponseType()

	requireScope := func() error {
		if len(r.Scope()) == 0 {
			return protocol.NewError("invalid_request", "scope is required")
		}
		return nil
	}

	switch {
	case responseType.Equal(SettingsActonResponseType):
		if r.SettingsAction() == settingsaction.SettingsActionDeleteAccount {
			if !h.AccountDeletionConfig.ScheduledByEndUserEnabled {
				return protocol.NewError("invalid_request", "account deletion by end user is disabled")
			}
		}
		fallthrough
	case responseType.Equal(CodeResponseType):
		if err := requireScope(); err != nil {
			return err
		}
		if client.IsPublic() {
			if r.CodeChallenge() == "" {
				return protocol.NewError("invalid_request", "PKCE code challenge is required for public clients")
			}
		}
		if r.CodeChallenge() != "" && r.CodeChallengeMethod() != pkce.CodeChallengeMethodS256 {
			return protocol.NewError("invalid_request", "only 'S256' PKCE transform is supported")
		}
	case responseType.Equal(NoneResponseType):
		if err := requireScope(); err != nil {
			return err
		}
	case responseType.Equal(PreAuthenticatedURLTokenResponseType):
		if err := h.validatePreAuthenticatedURLTokenRequest(client, r); err != nil {
			return err
		}
	default:
		return protocol.NewError("unsupported_response_type", fmt.Sprintf("response_type: %v is not supported", responseType.Raw))
	}

	return nil
}

func (h *AuthorizationHandler) doValidateRequestWithoutTx(
	client *config.OAuthClientConfig,
	r protocol.AuthorizationRequest,
) error {
	if err := h.validateResponseTypeIsWhitelisted(r); err != nil {
		return err
	}

	if err := h.validatePrompt(r); err != nil {
		return err
	}

	if err := h.validateRequestParameters(client, r); err != nil {
		return err
	}

	if r.SSOEnabled() && client != nil && client.MaxConcurrentSession == 1 {
		return protocol.NewError("invalid_request", "'sso_enabled' must be false if config 'x_max_concurrent_session' is 1")
	}

	return nil
}

func (h *AuthorizationHandler) generateCodeResponse(
	ctx context.Context,
	createCodeGrantOptions *CreateCodeGrantOptions,
	resp protocol.AuthorizationResponse,
) error {
	code, _, err := h.CodeGrantService.CreateCodeGrant(ctx, createCodeGrantOptions)
	if err != nil {
		return err
	}

	resp.Code(code)
	return nil
}

func (h *AuthorizationHandler) generateSettingsActionResponse(
	ctx context.Context,
	redirectURI string,
	r protocol.AuthorizationRequest,
	resp protocol.AuthorizationResponse,
	userID string,
) error {
	code, _, err := h.SettingsActionGrantService.CreateSettingsActionGrant(ctx, &CreateSettingsActionGrantOptions{
		RedirectURI:          redirectURI,
		AuthorizationRequest: r,
		UserID:               userID,
	})
	if err != nil {
		return err
	}

	resp.Code(code)
	return nil
}

func (h *AuthorizationHandler) prepareConsentErrInvalidOAuthResponse(ctx context.Context, req *http.Request, oauthError protocol.OAuthProtocolError) httputil.Result {
	resultErr := AuthorizationResultError{
		Response: oauthError.Response,
	}

	state := req.URL.Query().Get("state")
	if state != "" {
		resultErr.Response.State(state)
	}

	_, client := resolveClient(ctx, h.ClientResolver, req.URL.Query().Get("client_id"))

	// Only redirect if oauth session is expired / not found
	// It mostly happens when user refresh the page or go back to the page after authentication
	if oauthError.Type() == "invalid_request" && client != nil {
		redirectURI, err := url.Parse(req.URL.Query().Get("redirect_uri"))
		if err == nil {
			err = validateRedirectURI(client, h.HTTPProto, h.HTTPOrigin, h.AppDomains, []string{}, redirectURI)
			if err == nil {
				resultErr.RedirectURI = redirectURI
			}
		}
	}

	return resultErr
}
