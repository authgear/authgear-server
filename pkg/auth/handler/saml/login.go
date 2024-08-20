package saml

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlbinding"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlerror"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol/samlprotocolhttp"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlsession"
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
				})
			h.handleErrorResponse(rw, r, errResponse)
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
				})
			h.handleErrorResponse(rw, r, errResponse)
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
	uiInfo, err := h.SAMLUIService.ResolveUIInfo(samlSessionEntry)
	if err != nil {
		panic(err)
	}

	// TODO(saml): Handle prompt = none case

	samlSession := samlsession.NewSAMLSession(samlSessionEntry, uiInfo)
	err = h.SAMLSessionService.Save(samlSession)
	if err != nil {
		panic(err)
	}

	endpoint, err := h.SAMLUIService.BuildAuthenticationURL(samlSession)
	if err != nil {
		panic(err)
	}

	resp := &httputil.ResultRedirect{
		URL: endpoint.String(),
	}
	resp.WriteResponse(rw, r)
}

func (h *LoginHandler) handleErrorResponse(
	rw http.ResponseWriter, r *http.Request,
	result *samlprotocolhttp.SAMLErrorResult,
) {
	h.Logger.WithError(result).Warnln("saml login failed with expected error")

	result.WriteResponse(rw, r)
}

func (h *LoginHandler) handleUnknownError(
	rw http.ResponseWriter, r *http.Request,
	callbackURL string,
	relayState string,
	err any,
) {
	now := h.Logger.Time.UTC()
	e := panicutil.MakeError(err)
	h.Logger.WithError(e).Error("panic occurred")

	result := samlprotocolhttp.NewSAMLErrorResult(e,
		samlprotocolhttp.SAMLResult{
			CallbackURL: callbackURL,
			Binding:     samlprotocol.SAMLBindingHTTPPost,
			Response:    samlprotocol.NewUnexpectedServerErrorResponse(now, h.SAMLService.IdpEntityID()),
			RelayState:  relayState,
		})
	result.WriteResponse(rw, r)
}
