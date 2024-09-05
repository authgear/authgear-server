package saml

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlbinding"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlerror"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol/samlprotocolhttp"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlsession"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/panicutil"
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

type LoginHandler struct {
	Logger             *LoginHandlerLogger
	Clock              clock.Clock
	SAMLConfig         *config.SAMLConfig
	SAMLService        HandlerSAMLService
	SAMLSessionService SAMLSessionService
	SAMLUIService      SAMLUIService

	LoginResultHandler LoginResultHandler
}

func (h *LoginHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	now := h.Clock.NowUTC()
	serviceProviderId := httproute.GetParam(r, "service_provider_id")
	sp, ok := h.SAMLConfig.ResolveProvider(serviceProviderId)
	if !ok {
		http.NotFound(rw, r)
		return
	}

	callbackURL := sp.DefaultAcsURL()
	issuer := h.SAMLService.IdpEntityID()
	var relayState string

	defer func() {
		if err := recover(); err != nil {
			h.handleUnknownError(rw, r, callbackURL, relayState, err)
		}
	}()

	var err error
	var parseResult samlbinding.SAMLBindingParseResult

	switch r.Method {
	case "GET":
		// HTTP-Redirect binding
		parseResult, err = samlbinding.SAMLBindingHTTPRedirectParse(r)
	case "POST":
		// HTTP-POST binding
		parseResult, err = samlbinding.SAMLBindingHTTPPostParse(r)
	default:
		panic(fmt.Errorf("unexpected method %s", r.Method))
	}

	if err != nil {
		if errors.Is(err, samlbinding.ErrNoRequest) {
			// This is a IdP-initated flow
			result := h.handleIdpInitiated(
				r.Context(),
				sp,
				callbackURL,
			)
			h.writeResult(rw, r, result)
			return
		}

		var parseRequestFailedErr *samlerror.ParseRequestFailedError
		if errors.As(err, &parseRequestFailedErr) {
			errResponse := samlprotocolhttp.NewExpectedSAMLErrorResult(err,
				samlprotocolhttp.SAMLResult{
					CallbackURL: callbackURL,
					Binding:     samlprotocol.SAMLBindingHTTPPost,
					Response: samlprotocol.NewRequestDeniedErrorResponse(
						now,
						issuer,
						"failed to parse SAMLRequest",
						parseRequestFailedErr.GetDetailElements(),
					),
				},
			)
			h.writeResult(rw, r, errResponse)
			return
		}
		panic(err)
	}

	var authnRequest *samlprotocol.AuthnRequest
	switch parseResult := parseResult.(type) {
	case *samlbinding.SAMLBindingHTTPRedirectParseResult:
		relayState = parseResult.RelayState
		err = h.SAMLService.ValidateAuthnRequest(sp.GetID(), parseResult.AuthnRequest)
		// TODO(saml): Validate the signature in parseResult
		authnRequest = parseResult.AuthnRequest
	case *samlbinding.SAMLBindingHTTPPostParseResult:
		relayState = parseResult.RelayState
		err = h.SAMLService.ValidateAuthnRequest(sp.GetID(), parseResult.AuthnRequest)
		// TODO(saml): Validate the signature in AuthnRequest
		authnRequest = parseResult.AuthnRequest
	default:
		panic("unexpected parse result type")
	}

	if err != nil {
		var invalidRequestErr *samlerror.InvalidRequestError
		if errors.As(err, &invalidRequestErr) {
			errResponse := samlprotocolhttp.NewExpectedSAMLErrorResult(err,
				samlprotocolhttp.SAMLResult{
					CallbackURL: callbackURL,
					Binding:     samlprotocol.SAMLBindingHTTPPost,
					Response: samlprotocol.NewRequestDeniedErrorResponse(
						now,
						issuer,
						fmt.Sprintf("invalid SAMLRequest: %s", invalidRequestErr.Reason),
						invalidRequestErr.GetDetailElements(),
					),
					RelayState: relayState,
				},
			)
			h.writeResult(rw, r, errResponse)
			return
		}
		panic(err)
	}

	if authnRequest.AssertionConsumerServiceURL != "" {
		callbackURL = authnRequest.AssertionConsumerServiceURL
	}

	result := h.startSSOFlow(
		r.Context(),
		sp,
		string(authnRequest.ToXMLBytes()),
		callbackURL,
		relayState,
		authnRequest.GetIsPassive(),
	)
	h.writeResult(rw, r, result)
}

func (h *LoginHandler) handleUnknownError(
	rw http.ResponseWriter, r *http.Request,
	callbackURL string,
	relayState string,
	err any,
) {
	now := h.Logger.Time.UTC()
	e := panicutil.MakeError(err)

	result := samlprotocolhttp.NewUnexpectedSAMLErrorResult(e,
		samlprotocolhttp.SAMLResult{
			CallbackURL: callbackURL,
			Binding:     samlprotocol.SAMLBindingHTTPPost,
			Response:    samlprotocol.NewUnexpectedServerErrorResponse(now, h.SAMLService.IdpEntityID()),
			RelayState:  relayState,
		},
	)
	h.writeResult(rw, r, result)
}

func (h *LoginHandler) writeResult(
	rw http.ResponseWriter, r *http.Request,
	result httputil.Result,
) {
	switch result := result.(type) {
	case *samlprotocolhttp.SAMLErrorResult:
		if result.IsUnexpected {
			h.Logger.WithError(result.Cause).Error("unexpected error")
		} else {
			h.Logger.WithError(result).Warnln("saml login failed with expected error")
		}
	}
	result.WriteResponse(rw, r)
}

func (h *LoginHandler) handleIdpInitiated(
	ctx context.Context,
	sp *config.SAMLServiceProviderConfig,
	callbackURL string,
) httputil.Result {
	return h.startSSOFlow(
		ctx,
		sp,
		"",
		callbackURL,
		"",
		false,
	)
}

func (h *LoginHandler) startSSOFlow(
	ctx context.Context,
	sp *config.SAMLServiceProviderConfig,
	authnRequestXML string,
	callbackURL string,
	relayState string,
	isPassive bool,
) httputil.Result {
	now := h.Clock.NowUTC()
	issuer := h.SAMLService.IdpEntityID()

	samlSessionEntry := &samlsession.SAMLSessionEntry{
		ServiceProviderID: sp.GetID(),
		AuthnRequestXML:   authnRequestXML,
		CallbackURL:       callbackURL,
		RelayState:        relayState,
	}

	if isPassive == true {
		// If IsPassive=true, no ui should be displayed.
		// Authenticate by existing session or error.
		var resolvedSession session.ResolvedSession
		if s := session.GetSession(ctx); s != nil {
			resolvedSession = s
		}
		// Ignore any session that is not allow to be used here
		if !oauth.ContainsAllScopes(oauth.SessionScopes(resolvedSession), []string{oauth.PreAuthenticatedURLScope}) {
			resolvedSession = nil
		}

		if resolvedSession == nil {
			// No session, return NoPassive error.
			err := fmt.Errorf("no session but IsPassive=true")
			result := samlprotocolhttp.NewExpectedSAMLErrorResult(err,
				samlprotocolhttp.SAMLResult{
					CallbackURL: callbackURL,
					Binding:     samlprotocol.SAMLBindingHTTPPost,
					Response: samlprotocol.NewNoPassiveErrorResponse(
						now,
						issuer,
					),
					RelayState: relayState,
				},
			)
			return result
		} else {
			// Else, authenticate with the existing session.
			authInfo := resolvedSession.CreateNewAuthenticationInfoByThisSession()
			// TODO(saml): If <Subject> is provided in the request,
			// ensure the user of current session matches the subject.
			result := h.LoginResultHandler.handleLoginResult(&authInfo, samlSessionEntry)
			return result
		}

	}

	uiInfo, err := h.SAMLUIService.ResolveUIInfo(samlSessionEntry)
	if err != nil {
		panic(err)
	}

	samlSession := samlsession.NewSAMLSession(samlSessionEntry, uiInfo)
	err = h.SAMLSessionService.Save(samlSession)
	if err != nil {
		panic(err)
	}

	endpoint, err := h.SAMLUIService.BuildAuthenticationURL(samlSession)
	if err != nil {
		panic(err)
	}

	result := &httputil.ResultRedirect{
		URL: endpoint.String(),
	}

	return result
}
