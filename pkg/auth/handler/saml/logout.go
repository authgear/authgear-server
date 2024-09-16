package saml

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlbinding"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlerror"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol/samlprotocolhttp"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/panicutil"
)

func ConfigureLogoutRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST").
		WithPathPattern("/saml2/logout/:service_provider_id")
}

type LogoutHandlerLogger struct{ *log.Logger }

func NewLogoutHandlerLogger(lf *log.Factory) *LogoutHandlerLogger {
	return &LogoutHandlerLogger{lf.New("saml-logout-handler")}
}

type LogoutHandler struct {
	Logger      *LogoutHandlerLogger
	Clock       clock.Clock
	SAMLConfig  *config.SAMLConfig
	SAMLService HandlerSAMLService
}

func (h *LogoutHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	serviceProviderId := httproute.GetParam(r, "service_provider_id")
	sp, ok := h.SAMLConfig.ResolveProvider(serviceProviderId)
	if !ok {
		http.NotFound(rw, r)
		return
	}

	if !sp.SLOEnabled {
		http.NotFound(rw, r)
		return
	}

	callbackURL := sp.SLOCallbackURL
	responseBinding := sp.SLOBinding

	defer func() {
		if err := recover(); err != nil {
			h.handleUnknownError(rw, r, callbackURL, responseBinding, "", err)
		}
	}()

	result := h.handleSLORequest(r, sp, callbackURL, responseBinding)
	h.writeResult(rw, r, result)
}

func (h *LogoutHandler) handleSLORequest(
	r *http.Request,
	sp *config.SAMLServiceProviderConfig,
	callbackURL string,
	responseBinding samlprotocol.SAMLBinding,
) httputil.Result {
	now := h.Logger.Time.UTC()
	var parseResult samlbinding.SAMLBindingParseResult
	var relayState string
	var logoutRequest *samlprotocol.LogoutRequest

	// Get data with corresponding binding
	var err error
	switch r.Method {
	case "GET":
		// HTTP-Redirect binding
		r, e := samlbinding.SAMLBindingHTTPRedirectParse(r)
		if e != nil {
			err = e
			break
		}
		logoutRequest, err = samlprotocol.ParseLogoutRequest([]byte(r.SAMLRequestXML))
		if err != nil {
			err = &samlerror.ParseRequestFailedError{
				Reason: "malformed LogoutRequest",
				Cause:  err,
			}
			break
		}
		parseResult = r
		relayState = r.RelayState
	case "POST":
		// HTTP-POST binding
		r, e := samlbinding.SAMLBindingHTTPPostParse(r)
		if e != nil {
			err = e
			break
		}
		logoutRequest, err = samlprotocol.ParseLogoutRequest([]byte(r.SAMLRequestXML))
		if err != nil {
			err = &samlerror.ParseRequestFailedError{
				Reason: "malformed LogoutRequest",
				Cause:  err,
			}
			break
		}
		parseResult = r
		relayState = r.RelayState
	default:
		// panic because it should not happen if ConfigureLogoutRoute is correct
		panic("unexpected method")
	}

	if err != nil {
		var parseRequestFailedErr *samlerror.ParseRequestFailedError
		if errors.As(err, &parseRequestFailedErr) {
			return samlprotocolhttp.NewExpectedSAMLErrorResult(err,
				samlprotocolhttp.SAMLResult{
					CallbackURL: callbackURL,
					Binding:     responseBinding,
					Response: samlprotocol.NewRequestDeniedErrorResponse(
						now,
						h.SAMLService.IdpEntityID(),
						"failed to parse SAMLRequest",
						parseRequestFailedErr.GetDetailElements(),
					),
				},
			)
		} else {
			return h.makeUnknownErrorResult(
				callbackURL,
				responseBinding,
				relayState,
				err,
			)
		}
	}

	// Verify the signature
	switch parseResult := parseResult.(type) {
	case *samlbinding.SAMLBindingHTTPRedirectParseResult:
		err = h.SAMLService.VerifyExternalSignature(sp,
			parseResult.SAMLRequest,
			parseResult.SigAlg,
			parseResult.RelayState,
			parseResult.Signature)
		if err != nil {
			break
		}
	case *samlbinding.SAMLBindingHTTPPostParseResult:
		relayState = parseResult.RelayState
		err = h.SAMLService.VerifyEmbeddedSignature(sp, parseResult.SAMLRequestXML)
		if err != nil {
			break
		}
	default:
		panic("unexpected parse result type")
	}
	if err != nil {
		var invalidSignatureErr *samlerror.InvalidSignatureError
		if errors.As(err, &invalidSignatureErr) {
			return samlprotocolhttp.NewExpectedSAMLErrorResult(err,
				samlprotocolhttp.SAMLResult{
					CallbackURL: callbackURL,
					Binding:     samlprotocol.SAMLBindingHTTPPost,
					Response: samlprotocol.NewRequestDeniedErrorResponse(
						now,
						h.SAMLService.IdpEntityID(),
						"invalid signature",
						invalidSignatureErr.GetDetailElements(),
					),
					RelayState: relayState,
				},
			)
		} else {
			return h.makeUnknownErrorResult(
				callbackURL,
				responseBinding,
				relayState,
				err,
			)
		}
	}

	response, err := h.SAMLService.IssueLogoutResponse(
		callbackURL,
		sp.GetID(),
		logoutRequest,
	)
	if err != nil {
		return h.makeUnknownErrorResult(
			callbackURL,
			responseBinding,
			relayState,
			err,
		)
	}

	return &samlprotocolhttp.SAMLResult{
		CallbackURL: callbackURL,
		Binding:     responseBinding,
		Response:    response,
		RelayState:  relayState,
		// TODO(tung): Refactor the code to inject the signer automatically
		Signer: h.SAMLService,
	}
}

func (h *LogoutHandler) makeUnknownErrorResult(
	callbackURL string,
	responseBinding samlprotocol.SAMLBinding,
	relayState string,
	err any,
) *samlprotocolhttp.SAMLErrorResult {
	now := h.Logger.Time.UTC()
	e := panicutil.MakeError(err)

	return samlprotocolhttp.NewUnexpectedSAMLErrorResult(e,
		samlprotocolhttp.SAMLResult{
			CallbackURL: callbackURL,
			Binding:     responseBinding,
			Response:    samlprotocol.NewUnexpectedServerErrorResponse(now, h.SAMLService.IdpEntityID()),
			RelayState:  relayState,
		},
	)

}

func (h *LogoutHandler) handleUnknownError(
	rw http.ResponseWriter, r *http.Request,
	callbackURL string,
	responseBinding samlprotocol.SAMLBinding,
	relayState string,
	err any,
) {
	result := h.makeUnknownErrorResult(
		callbackURL,
		responseBinding,
		relayState,
		err,
	)
	h.writeResult(rw, r, result)
}

func (h *LogoutHandler) writeResult(
	rw http.ResponseWriter, r *http.Request,
	result httputil.Result,
) {
	switch result := result.(type) {
	case *samlprotocolhttp.SAMLErrorResult:
		if result.IsUnexpected {
			h.Logger.WithError(result.Cause).Error("unexpected error")
		} else {
			h.Logger.WithError(result).Warnln("saml logout failed with expected error")
		}
	}
	result.WriteResponse(rw, r)
}
