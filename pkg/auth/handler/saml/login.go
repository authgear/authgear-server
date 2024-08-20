package saml

import (
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
		var parseRequestFailedErr *samlerror.ParseRequestFailedError
		if errors.As(err, &parseRequestFailedErr) {
			errResponse := samlprotocolhttp.NewSAMLErrorResult(err,
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
				false,
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
		err = h.SAMLService.ValidateAuthnRequest(sp.ID, parseResult.AuthnRequest)
		// TODO(saml): Validate the signature in parseResult
		authnRequest = parseResult.AuthnRequest
	case *samlbinding.SAMLBindingHTTPPostParseResult:
		relayState = parseResult.RelayState
		err = h.SAMLService.ValidateAuthnRequest(sp.ID, parseResult.AuthnRequest)
		// TODO(saml): Validate the signature in AuthnRequest
		authnRequest = parseResult.AuthnRequest
	default:
		panic("unexpected parse result type")
	}

	if err != nil {
		var invalidRequestErr *samlerror.InvalidRequestError
		if errors.As(err, &invalidRequestErr) {
			errResponse := samlprotocolhttp.NewSAMLErrorResult(err,
				samlprotocolhttp.SAMLResult{
					CallbackURL: callbackURL,
					Binding:     samlprotocol.SAMLBindingHTTPPost,
					Response: samlprotocol.NewRequestDeniedErrorResponse(
						now,
						issuer,
						"invalid SAMLRequest",
						invalidRequestErr.GetDetailElements(),
					),
					RelayState: relayState,
				},
				false,
			)
			h.writeResult(rw, r, errResponse)
			return
		}
		panic(err)
	}

	if authnRequest.AssertionConsumerServiceURL != "" {
		callbackURL = authnRequest.AssertionConsumerServiceURL
	}

	samlSessionEntry := &samlsession.SAMLSessionEntry{
		ServiceProviderID: sp.ID,
		AuthnRequestXML:   string(authnRequest.ToXMLBytes()),
		CallbackURL:       callbackURL,
		RelayState:        relayState,
	}

	if authnRequest.GetIsPassive() == true {
		// If IsPassive=true, no ui should be displayed.
		// Authenticate by existing session or error.
		var resolvedSession session.ResolvedSession
		if s := session.GetSession(r.Context()); s != nil {
			resolvedSession = s
		}
		// Ignore any session that is not allow to be used here
		if !oauth.ContainsAllScopes(oauth.SessionScopes(resolvedSession), []string{oauth.PreAuthenticatedURLScope}) {
			resolvedSession = nil
		}

		if resolvedSession == nil {
			// No session, return NoPassive error.
			errResponse := samlprotocolhttp.NewSAMLErrorResult(err,
				samlprotocolhttp.SAMLResult{
					CallbackURL: callbackURL,
					Binding:     samlprotocol.SAMLBindingHTTPPost,
					Response: samlprotocol.NewNoPassiveErrorResponse(
						now,
						issuer,
					),
					RelayState: relayState,
				},
				false,
			)
			h.writeResult(rw, r, errResponse)
			return
		} else {
			// Else, authenticate with the existing session.
			authInfo := resolvedSession.CreateNewAuthenticationInfoByThisSession()
			// TODO(saml): If <Subject> is provided in the request,
			// ensure the user of current session matches the subject.
			result := h.LoginResultHandler.handleLoginResult(&authInfo, samlSessionEntry)
			h.writeResult(rw, r, result)
			return
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

	result := samlprotocolhttp.NewSAMLErrorResult(e,
		samlprotocolhttp.SAMLResult{
			CallbackURL: callbackURL,
			Binding:     samlprotocol.SAMLBindingHTTPPost,
			Response:    samlprotocol.NewUnexpectedServerErrorResponse(now, h.SAMLService.IdpEntityID()),
			RelayState:  relayState,
		},
		true,
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
