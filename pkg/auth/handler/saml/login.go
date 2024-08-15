package saml

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/saml"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlbinding"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
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

	defer func() {
		if err := recover(); err != nil {
			h.handleUnknownError(rw, callbackURL, err)
		}
	}()

	var err error
	var relayState string
	var authnRequest *saml.AuthnRequest

	switch r.Method {
	case "GET":
		// HTTP-Redirect binding
		authnRequest, relayState, err = h.handleRedirectBinding(now, r)
	case "POST":
		// HTTP-POST binding
		authnRequest, relayState, err = h.handlePostBinding(now, r)
	default:
		panic(fmt.Errorf("unexpected method %s", r.Method))
	}

	if err != nil {
		var protocolErr *samlprotocol.SAMLProtocolError
		if errors.As(err, &protocolErr) {
			h.handleProtocolError(rw, callbackURL, protocolErr)
			return
		}
		panic(err)
	}

	err = h.SAMLService.ValidateAuthnRequest(sp.ID, authnRequest)
	if err != nil {
		protocolErr := &samlprotocol.SAMLProtocolError{
			Response:   saml.NewRequestDeniedErrorResponse(now, fmt.Sprintf("invalid SAMLRequest: %v", err)),
			RelayState: relayState,
			Cause:      err,
		}
		h.handleProtocolError(rw, callbackURL, protocolErr)
		return
	}

	if authnRequest.AssertionConsumerServiceURL != "" {
		callbackURL = authnRequest.AssertionConsumerServiceURL
	}

	samlSessionEntry := &samlsession.SAMLSessionEntry{
		ServiceProviderID: sp.ID,
		AuthnRequestXML:   string(authnRequest.ToXMLBytes()),
		CallbackURL:       callbackURL,
	}
	uiInfo, err := h.SAMLUIService.ResolveUIInfo(samlSessionEntry)
	if err != nil {
		protocolErr := &samlprotocol.SAMLProtocolError{
			Response:   saml.NewRequestDeniedErrorResponse(now, fmt.Sprintf("invalid SAMLRequest: %v", err)),
			RelayState: relayState,
			Cause:      err,
		}
		h.handleProtocolError(rw, callbackURL, protocolErr)
		return
	}

	// TODO(saml): Handle prompt = none case

	samlSession := samlsession.NewSAMLSession(samlSessionEntry, uiInfo)
	err = h.SAMLSessionService.Save(samlSession)
	if err != nil {
		panic(err)
	}

	endpoint, err := h.SAMLUIService.BuildAuthenticationURL(samlSession)

	resp := &httputil.ResultRedirect{
		URL: endpoint.String(),
	}
	resp.WriteResponse(rw, r)
}

func (h *LoginHandler) handleRedirectBinding(now time.Time, r *http.Request) (authnRequest *saml.AuthnRequest, relayState string, err error) {
	relayState = r.URL.Query().Get("RelayState")
	compressedRequest, err := base64.StdEncoding.DecodeString(r.URL.Query().Get("SAMLRequest"))
	if err != nil {
		return nil, relayState, &samlprotocol.SAMLProtocolError{
			Response:   saml.NewRequestDeniedErrorResponse(now, "failed to decode SAMLRequest"),
			RelayState: relayState,
			Cause:      err,
		}
	}
	requestBuffer, err := io.ReadAll(samlbinding.NewSaferFlateReader(bytes.NewReader(compressedRequest)))
	if err != nil {
		return nil, relayState, &samlprotocol.SAMLProtocolError{
			Response:   saml.NewRequestDeniedErrorResponse(now, "failed to decompress SAMLRequest"),
			RelayState: relayState,
			Cause:      err,
		}
	}

	authnRequest, err = saml.ParseAuthnRequest(requestBuffer)
	if err != nil {
		return nil, relayState, &samlprotocol.SAMLProtocolError{
			Response:   saml.NewRequestDeniedErrorResponse(now, "failed to parse SAMLRequest"),
			RelayState: relayState,
			Cause:      err,
		}
	}
	return authnRequest, relayState, nil
}

func (h *LoginHandler) handlePostBinding(now time.Time, r *http.Request) (authnRequest *saml.AuthnRequest, relayState string, err error) {
	if err := r.ParseForm(); err != nil {
		return nil, "", &samlprotocol.SAMLProtocolError{
			Response: saml.NewRequestDeniedErrorResponse(now, "failed to parse request body"),
			Cause:    err,
		}
	}
	relayState = r.PostForm.Get("RelayState")

	requestBuffer, err := base64.StdEncoding.DecodeString(r.PostForm.Get("SAMLRequest"))
	if err != nil {
		return nil, relayState, &samlprotocol.SAMLProtocolError{
			Response:   saml.NewRequestDeniedErrorResponse(now, "failed to decode SAMLRequest"),
			RelayState: relayState,
			Cause:      err,
		}
	}

	authnRequest, err = saml.ParseAuthnRequest(requestBuffer)
	if err != nil {
		return nil, relayState, &samlprotocol.SAMLProtocolError{
			Response:   saml.NewRequestDeniedErrorResponse(now, "failed to parse SAMLRequest"),
			RelayState: relayState,
			Cause:      err,
		}
	}
	return authnRequest, relayState, nil
}

func (h *LoginHandler) handleProtocolError(rw http.ResponseWriter, callbackURL string, err *samlprotocol.SAMLProtocolError) {
	h.Logger.Warnln(err.Error())
	// TODO(saml): Return the error to callbackURL
	panic(err)
}

func (h *LoginHandler) handleUnknownError(rw http.ResponseWriter, callbackURL string, err any) {
	e := panicutil.MakeError(err)
	h.Logger.WithError(e).Error("panic occurred")
	// TODO(saml): Return a error response to callbackURL
	panic(err)
}
