package saml

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlbinding"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
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

	BindingHTTPPostWriter     BindingHTTPPostWriter
	BindingHTTPRedirectWriter BindingHTTPRedirectWriter
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
	var relayState string

	defer func() {
		if err := recover(); err != nil {
			e := panicutil.MakeError(err)
			h.handleError(rw, r, responseBinding, callbackURL, relayState, e)
		}
	}()

	var response samlprotocol.Respondable
	var err error
	relayState, response, err = h.handleSLORequest(r, sp, callbackURL)
	if err != nil {
		h.handleError(rw, r, responseBinding, callbackURL, relayState, err)
		return
	}
	h.writeResponse(rw, r, response, responseBinding, callbackURL, relayState)
}

func (h *LogoutHandler) parseSLORequest(
	r *http.Request,
) (
	parseResult samlbinding.SAMLBindingParseResult,
	logoutRequest *samlprotocol.LogoutRequest,
	relayState string,
	err error,
) {
	defer func() {
		// Transform known errors
		if err != nil {
			var parseRequestFailedErr *samlprotocol.ParseRequestFailedError
			if errors.As(err, &parseRequestFailedErr) {
				err = NewSAMLErrorResult(err,
					samlprotocol.NewRequestDeniedErrorResponse(
						h.Clock.NowUTC(),
						h.SAMLService.IdpEntityID(),
						"failed to parse SAMLRequest",
						parseRequestFailedErr.GetDetailElements(),
					),
				)
			}
		}
	}()
	switch r.Method {
	case "GET":
		// HTTP-Redirect binding
		r, parseErr := samlbinding.SAMLBindingHTTPRedirectParse(r)
		if parseErr != nil {
			return nil, nil, "", parseErr
		}
		logoutRequest, err = samlprotocol.ParseLogoutRequest([]byte(r.SAMLRequestXML))
		if err != nil {
			err = &samlprotocol.ParseRequestFailedError{
				Reason: "malformed LogoutRequest",
				Cause:  err,
			}
			return nil, nil, "", err
		}
		parseResult = r
		relayState = r.RelayState
	case "POST":
		// HTTP-POST binding
		r, parseErr := samlbinding.SAMLBindingHTTPPostParse(r)
		if parseErr != nil {
			return nil, nil, "", parseErr
		}
		logoutRequest, err = samlprotocol.ParseLogoutRequest([]byte(r.SAMLRequestXML))
		if err != nil {
			err = &samlprotocol.ParseRequestFailedError{
				Reason: "malformed LogoutRequest",
				Cause:  err,
			}
			return nil, nil, "", err
		}
		parseResult = r
		relayState = r.RelayState
	default:
		// panic because it should not happen if ConfigureLogoutRoute is correct
		panic("unexpected method")
	}
	return parseResult, logoutRequest, relayState, nil
}

func (h *LogoutHandler) verifySignature(
	sp *config.SAMLServiceProviderConfig,
	parseResult samlbinding.SAMLBindingParseResult,
) (err error) {
	defer func() {
		// Transform known errors
		if err != nil {
			var invalidSignatureErr *samlprotocol.InvalidSignatureError
			if errors.As(err, &invalidSignatureErr) {
				err = NewSAMLErrorResult(err,
					samlprotocol.NewRequestDeniedErrorResponse(
						h.Clock.NowUTC(),
						h.SAMLService.IdpEntityID(),
						"invalid signature",
						invalidSignatureErr.GetDetailElements(),
					),
				)
			}
		}
	}()
	switch parseResult := parseResult.(type) {
	case *samlbinding.SAMLBindingHTTPRedirectParseResult:
		err = h.SAMLService.VerifyExternalSignature(sp,
			parseResult.SAMLRequest,
			parseResult.SigAlg,
			parseResult.RelayState,
			parseResult.Signature)
		if err != nil {
			return err
		}
	case *samlbinding.SAMLBindingHTTPPostParseResult:
		err = h.SAMLService.VerifyEmbeddedSignature(sp, parseResult.SAMLRequestXML)
		if err != nil {
			return err
		}
	default:
		panic("unexpected parse result type")
	}
	return nil
}

func (h *LogoutHandler) handleSLORequest(
	r *http.Request,
	sp *config.SAMLServiceProviderConfig,
	callbackURL string,
) (relayState string, response samlprotocol.Respondable, err error) {
	var parseResult samlbinding.SAMLBindingParseResult
	var logoutRequest *samlprotocol.LogoutRequest

	// Get data with corresponding binding
	parseResult, logoutRequest, relayState, err = h.parseSLORequest(r)
	if err != nil {
		return relayState, nil, err
	}

	// Verify the signature
	err = h.verifySignature(sp, parseResult)
	if err != nil {
		return relayState, nil, err
	}

	response, err = h.SAMLService.IssueLogoutResponse(
		callbackURL,
		sp.GetID(),
		logoutRequest,
	)
	if err != nil {
		return relayState, nil, err
	}

	return relayState, response, nil
}

func (h *LogoutHandler) writeResponse(
	rw http.ResponseWriter, r *http.Request,
	response samlprotocol.Respondable,
	responseBinding samlprotocol.SAMLBinding,
	callbackURL string,
	relayState string,
) {
	switch responseBinding {
	case samlprotocol.SAMLBindingHTTPPost:
		err := h.BindingHTTPPostWriter.Write(rw, r,
			callbackURL,
			response,
			relayState,
		)
		if err != nil {
			panic(err)
		}
	case samlprotocol.SAMLBindingHTTPRedirect:
		err := h.BindingHTTPRedirectWriter.Write(rw, r,
			callbackURL,
			response,
			relayState,
		)
		if err != nil {
			panic(err)
		}
	}
}

func (h *LogoutHandler) handleError(
	rw http.ResponseWriter,
	r *http.Request,
	responseBinding samlprotocol.SAMLBinding,
	callbackURL string,
	relayState string,
	err error,
) {
	now := h.Clock.NowUTC()
	var samlErrResult *SAMLErrorResult
	var response samlprotocol.Respondable
	if errors.As(err, &samlErrResult) {
		h.Logger.WithError(samlErrResult.Cause).Warnln("saml logout failed with expected error")
		response = samlErrResult.Response
	} else {
		h.Logger.WithError(err).Error("unexpected error")
		response = samlprotocol.NewUnexpectedServerErrorResponse(now, h.SAMLService.IdpEntityID())
	}
	h.writeResponse(rw, r, response, responseBinding, callbackURL, relayState)
}
