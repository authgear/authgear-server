package saml

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlbinding"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlsession"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/panicutil"
	"github.com/authgear/authgear-server/pkg/util/setutil"
)

func ConfigureLoginRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST").
		WithPathPattern("/saml2/login/:service_provider_id")
}

type LoginHandlerLogger struct{ *log.Logger }

func NewLoginHandlerLogger(lf *log.Factory) *LoginHandlerLogger {
	return &LoginHandlerLogger{lf.New("saml-login-handler")}
}

type loginResult interface {
	loginResult()
}

type loginResultRedirect struct {
	RedirectURL string
}

var _ loginResult = &loginResultRedirect{}

func (l *loginResultRedirect) loginResult() {}

type loginResultSAMLResponse struct {
	Response    samlprotocol.Respondable
	CallbackURL string
	RelayState  string
}

var _ loginResult = &loginResultSAMLResponse{}

func (l *loginResultSAMLResponse) loginResult() {}

type LoginHandler struct {
	Logger             *LoginHandlerLogger
	Clock              clock.Clock
	Database           *appdb.Handle
	SAMLConfig         *config.SAMLConfig
	SAMLService        HandlerSAMLService
	SAMLSessionService SAMLSessionService
	SAMLUIService      SAMLUIService

	UserFacade SAMLUserFacade

	LoginResultHandler    LoginResultHandler
	BindingHTTPPostWriter BindingHTTPPostWriter
}

func (h *LoginHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	serviceProviderId := httproute.GetParam(r, "service_provider_id")
	sp, ok := h.SAMLConfig.ResolveProvider(serviceProviderId)
	if !ok {
		http.NotFound(rw, r)
		return
	}

	callbackURL := sp.DefaultAcsURL()
	var relayState string

	defer func() {
		if err := recover(); err != nil {
			e := panicutil.MakeError(err)
			h.handleError(rw, r, callbackURL, relayState, e)
		}
	}()

	var result loginResult
	var err error
	relayState, callbackURL, result, err = h.handleLoginRequest(rw, r, sp, callbackURL)
	if err != nil {
		h.handleError(rw, r, callbackURL, relayState, err)
		return
	}
	switch result := result.(type) {
	case *loginResultRedirect:
		redirectResult := &httputil.ResultRedirect{
			URL: result.RedirectURL,
		}
		redirectResult.WriteResponse(rw, r)
		return
	case *loginResultSAMLResponse:
		if result.CallbackURL != "" {
			callbackURL = result.CallbackURL
		}
		err := h.BindingHTTPPostWriter.WriteResponse(rw, r,
			callbackURL,
			result.Response.Element(),
			result.RelayState,
		)
		if err != nil {
			panic(err)
		}
		return
	}
}

func (h *LoginHandler) parseRequest(r *http.Request,
) (
	result samlbinding.SAMLBindingParseReqeustResult,
	authnRequest *samlprotocol.AuthnRequest,
	relayState string,
	err error,
) {
	defer func() {
		// Transform known errors
		if err != nil {
			var parseRequestFailedErr *samlprotocol.ParseRequestFailedError
			if errors.As(err, &parseRequestFailedErr) {
				now := h.Clock.NowUTC()
				issuer := h.SAMLService.IdpEntityID()
				err = NewSAMLErrorResult(err,
					samlprotocol.NewRequestDeniedErrorResponse(
						now,
						issuer,
						"failed to parse SAMLRequest",
						parseRequestFailedErr.GetDetailElements(),
					),
				)
				return
			}
		}
	}()

	// Get data with corresponding binding
	switch r.Method {
	case "GET":
		// HTTP-Redirect binding
		r, parseErr := samlbinding.SAMLBindingHTTPRedirectParseRequest(r)
		if parseErr != nil {
			return nil, nil, "", parseErr
		}
		relayState = r.RelayState
		authnRequest, err = samlprotocol.ParseAuthnRequest([]byte(r.SAMLRequestXML))
		if err != nil {
			cause := err
			err = &samlprotocol.ParseRequestFailedError{
				Reason: "malformed AuthnRequest",
				Cause:  cause,
			}
			return
		}
		result = r
	case "POST":
		// HTTP-POST binding
		r, parseErr := samlbinding.SAMLBindingHTTPPostParseRequest(r)
		if parseErr != nil {
			return nil, nil, "", parseErr
		}
		relayState = r.RelayState
		authnRequest, err = samlprotocol.ParseAuthnRequest([]byte(r.SAMLRequestXML))
		if err != nil {
			cause := err
			err = &samlprotocol.ParseRequestFailedError{
				Reason: "malformed AuthnRequest",
				Cause:  cause,
			}
			return
		}
		result = r
	default:
		panic(fmt.Errorf("unexpected method %s", r.Method))
	}
	return result, authnRequest, relayState, nil
}

func (h *LoginHandler) verifyRequestSignature(
	sp *config.SAMLServiceProviderConfig,
	parseResult samlbinding.SAMLBindingParseReqeustResult,
) (err error) {
	defer func() {
		// Transform known errors
		if err != nil {
			var invalidSignatureErr *samlprotocol.InvalidSignatureError
			if errors.As(err, &invalidSignatureErr) {
				now := h.Clock.NowUTC()
				issuer := h.SAMLService.IdpEntityID()
				err = NewSAMLErrorResult(err,
					samlprotocol.NewRequestDeniedErrorResponse(
						now,
						issuer,
						"invalid signature",
						invalidSignatureErr.GetDetailElements(),
					),
				)
			}
		}
	}()

	// Verify the signature
	switch parseResult := parseResult.(type) {
	case *samlbinding.SAMLBindingHTTPRedirectParseRequestResult:
		err = h.SAMLService.VerifyExternalSignature(sp,
			parseResult.SAMLRequest,
			parseResult.SigAlg,
			parseResult.RelayState,
			parseResult.Signature)
		if err != nil {
			return err
		}
	case *samlbinding.SAMLBindingHTTPPostParseRequestResult:
		err = h.SAMLService.VerifyEmbeddedSignature(sp, parseResult.SAMLRequestXML)
		if err != nil {
			return err
		}
	default:
		panic("unexpected parse result type")
	}
	return nil
}

func (h *LoginHandler) handleLoginRequest(
	rw http.ResponseWriter,
	r *http.Request,
	sp *config.SAMLServiceProviderConfig,
	defaultCallbackURL string,
) (relayState string, callbackURL string, result loginResult, err error) {
	now := h.Clock.NowUTC()
	var parseResult samlbinding.SAMLBindingParseReqeustResult
	var authnRequest *samlprotocol.AuthnRequest
	callbackURL = defaultCallbackURL
	issuer := h.SAMLService.IdpEntityID()

	parseResult, authnRequest, relayState, err = h.parseRequest(r)
	if err != nil {
		if errors.Is(err, samlbinding.ErrNoRequest) {
			// This is a IdP-initated flow
			result, err := h.handleIdpInitiated(
				r.Context(),
				sp,
				callbackURL,
			)
			return relayState, callbackURL, result, err
		} else {
			return relayState, callbackURL, nil, err
		}
	}

	// Verify the signature
	err = h.verifyRequestSignature(sp, parseResult)
	if err != nil {
		return relayState, callbackURL, nil, err
	}

	// Validate the AuthnRequest
	err = h.SAMLService.ValidateAuthnRequest(sp.GetID(), authnRequest)
	if err != nil {
		var invalidRequestErr *samlprotocol.InvalidRequestError
		if errors.As(err, &invalidRequestErr) {
			errorResult := NewSAMLErrorResult(err,
				samlprotocol.NewRequestDeniedErrorResponse(
					now,
					issuer,
					fmt.Sprintf("invalid AuthnRequest: %s", invalidRequestErr.Reason),
					invalidRequestErr.GetDetailElements(),
				),
			)
			return relayState, callbackURL, nil, errorResult
		} else {
			return relayState, callbackURL, nil, err
		}
	}

	if authnRequest.AssertionConsumerServiceURL != "" {
		callbackURL = authnRequest.AssertionConsumerServiceURL
	}

	result, err = h.startSSOFlow(
		r.Context(),
		sp,
		string(authnRequest.ToXMLBytes()),
		callbackURL,
		relayState,
	)
	return relayState, callbackURL, result, err
}

func (h *LoginHandler) handleIdpInitiated(
	ctx context.Context,
	sp *config.SAMLServiceProviderConfig,
	callbackURL string,
) (loginResult, error) {
	return h.startSSOFlow(
		ctx,
		sp,
		"",
		callbackURL,
		"",
	)
}

func (h *LoginHandler) finishWithoutUI(
	ctx context.Context,
	loginHint *oauth.LoginHint,
	samlSessionEntry *samlsession.SAMLSessionEntry,
) (result loginResult, err error) {
	var resolvedSession session.ResolvedSession
	if s := session.GetSession(ctx); s != nil {
		resolvedSession = s
	}
	// Ignore any session that is not allow to be used here
	if !oauth.ContainsAllScopes(oauth.SessionScopes(resolvedSession), []string{oauth.PreAuthenticatedURLScope}) {
		resolvedSession = nil
	}

	// Ignore any session that does not match login_hint
	err = h.Database.WithTx(func() error {
		if loginHint != nil && resolvedSession != nil {
			hintUserIDs, err := h.UserFacade.GetUserIDsByLoginHint(loginHint)
			if err != nil {
				return err
			}
			hintUserIDsSet := setutil.NewSetFromSlice(hintUserIDs, setutil.Identity[string])
			if !hintUserIDsSet.Has(resolvedSession.GetAuthenticationInfo().UserID) {
				resolvedSession = nil
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if resolvedSession == nil {
		// No session, return NoPassive error.
		err := fmt.Errorf("no session but IsPassive=true")
		errorResult := NewSAMLErrorResult(err,
			samlprotocol.NewNoPassiveErrorResponse(
				h.Clock.NowUTC(),
				h.SAMLService.IdpEntityID(),
			),
		)
		return nil, errorResult
	} else {
		// Else, authenticate with the existing session.
		authInfo := resolvedSession.CreateNewAuthenticationInfoByThisSession()
		response, err := h.LoginResultHandler.handleLoginResult(ctx, &authInfo, samlSessionEntry)
		if err != nil {
			return nil, err
		}
		return &loginResultSAMLResponse{
			Response:    response,
			RelayState:  samlSessionEntry.RelayState,
			CallbackURL: samlSessionEntry.CallbackURL,
		}, nil
	}
}

func (h *LoginHandler) startSSOFlow(
	ctx context.Context,
	sp *config.SAMLServiceProviderConfig,
	authnRequestXML string,
	callbackURL string,
	relayState string,
) (loginResult, error) {
	now := h.Clock.NowUTC()
	issuer := h.SAMLService.IdpEntityID()

	samlSessionEntry := &samlsession.SAMLSessionEntry{
		ServiceProviderID: sp.GetID(),
		AuthnRequestXML:   authnRequestXML,
		CallbackURL:       callbackURL,
		RelayState:        relayState,
	}

	uiInfo, showUI, err := h.SAMLUIService.ResolveUIInfo(sp, samlSessionEntry)
	if err != nil {
		var invalidRequestErr *samlprotocol.InvalidRequestError
		if errors.As(err, &invalidRequestErr) {
			errorResult := NewSAMLErrorResult(err,
				samlprotocol.NewRequestDeniedErrorResponse(
					now,
					issuer,
					fmt.Sprintf("invalid SAMLRequest: %s", invalidRequestErr.Reason),
					invalidRequestErr.GetDetailElements(),
				),
			)
			return nil, errorResult
		} else {
			return nil, err
		}
	}

	var loginHint *oauth.LoginHint
	l, err := oauth.ParseLoginHint(uiInfo.LoginHint)
	if err == nil {
		loginHint = l
	}

	if !showUI {
		// If IsPassive=true, no ui should be displayed.
		// Authenticate by existing session or error.
		return h.finishWithoutUI(ctx, loginHint, samlSessionEntry)
	}

	samlSession := samlsession.NewSAMLSession(samlSessionEntry, uiInfo)
	err = h.SAMLSessionService.Save(samlSession)
	if err != nil {
		return nil, err
	}

	endpoint, err := h.SAMLUIService.BuildAuthenticationURL(samlSession)
	if err != nil {
		return nil, err
	}

	return &loginResultRedirect{
		RedirectURL: endpoint.String(),
	}, nil
}

func (h *LoginHandler) handleError(
	rw http.ResponseWriter,
	r *http.Request,
	callbackURL string,
	relayState string,
	err error,
) {
	now := h.Clock.NowUTC()
	var samlErrResult *SAMLErrorResult
	if errors.As(err, &samlErrResult) {
		h.Logger.WithError(samlErrResult.Cause).Warnln("saml login failed with expected error")
		err = h.BindingHTTPPostWriter.WriteResponse(rw, r,
			callbackURL,
			samlErrResult.Response.Element(),
			relayState,
		)
		if err != nil {
			panic(err)
		}
	} else {
		h.Logger.WithError(err).Error("unexpected error")
		err = h.BindingHTTPPostWriter.WriteResponse(rw, r,
			callbackURL,
			samlprotocol.NewUnexpectedServerErrorResponse(now, h.SAMLService.IdpEntityID()).Element(),
			relayState,
		)
		if err != nil {
			panic(err)
		}
	}
}
