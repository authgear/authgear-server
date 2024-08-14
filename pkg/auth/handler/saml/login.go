package saml

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/saml"
	"github.com/authgear/authgear-server/pkg/lib/saml/binding"
	"github.com/authgear/authgear-server/pkg/lib/saml/protocol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
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
	Logger      *LoginHandlerLogger
	Clock       clock.Clock
	SAMLConfig  *config.SAMLConfig
	SAMLService HandlerSAMLService
}

func (h *LoginHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	serviceProviderId := httproute.GetParam(r, "service_provider_id")
	_, ok := h.SAMLConfig.ResolveProvider(serviceProviderId)
	if !ok {
		http.NotFound(rw, r)
		return
	}

	var err error
	var relayState string
	var authnRequest *saml.AuthnRequest

	switch r.Method {
	case "GET":
		// HTTP-Redirect binding
		authnRequest, relayState, err = h.handleRedirectBinding(serviceProviderId, r)
	case "POST":
		// HTTP-POST binding
		authnRequest, relayState, err = h.handlePostBinding(serviceProviderId, r)
	default:
		panic(fmt.Errorf("unexpected method %s", r.Method))
	}

	if err != nil {
		var protocolErr *protocol.SAMLProtocolError
		if errors.As(err, &protocolErr) {
			h.Logger.Warnln(protocolErr.Error())
			// TODO(saml): Return the error to acs url
		}
		panic(err)
	}

	// TODO(saml): Redirect to auth ui
	_, _ = rw.Write([]byte(authnRequest.ID + relayState))
}

func (h *LoginHandler) handleRedirectBinding(serviceProviderId string, r *http.Request) (authnRequest *saml.AuthnRequest, relayState string, err error) {
	now := h.Clock.NowUTC()
	compressedRequest, err := base64.StdEncoding.DecodeString(r.URL.Query().Get("SAMLRequest"))
	if err != nil {
		return nil, "", &protocol.SAMLProtocolError{
			Response: saml.NewRequestDeniedErrorResponse(now, "failed to decode SAMLRequest"),
			Cause:    err,
		}
	}
	requestBuffer, err := io.ReadAll(binding.NewSaferFlateReader(bytes.NewReader(compressedRequest)))
	if err != nil {
		return nil, "", &protocol.SAMLProtocolError{
			Response: saml.NewRequestDeniedErrorResponse(now, "failed to decompress SAMLRequest"),
			Cause:    err,
		}
	}
	relayState = r.URL.Query().Get("RelayState")

	authnRequest, err = h.SAMLService.ParseAuthnRequest(serviceProviderId, requestBuffer)
	if err != nil {
		return nil, "", &protocol.SAMLProtocolError{
			Response: saml.NewRequestDeniedErrorResponse(now, "failed to validate SAMLRequest"),
			Cause:    err,
		}
	}
	return authnRequest, relayState, nil
}

func (h *LoginHandler) handlePostBinding(serviceProviderId string, r *http.Request) (authnRequest *saml.AuthnRequest, relayState string, err error) {
	now := h.Clock.NowUTC()
	if err := r.ParseForm(); err != nil {
		return nil, "", &protocol.SAMLProtocolError{
			Response: saml.NewRequestDeniedErrorResponse(now, "failed to parse request body"),
			Cause:    err,
		}
	}

	requestBuffer, err := base64.StdEncoding.DecodeString(r.PostForm.Get("SAMLRequest"))
	if err != nil {
		return nil, "", &protocol.SAMLProtocolError{
			Response: saml.NewRequestDeniedErrorResponse(now, "failed to decode SAMLRequest"),
			Cause:    err,
		}
	}
	relayState = r.PostForm.Get("RelayState")

	authnRequest, err = h.SAMLService.ParseAuthnRequest(serviceProviderId, requestBuffer)
	if err != nil {
		return nil, "", &protocol.SAMLProtocolError{
			Response: saml.NewRequestDeniedErrorResponse(now, "failed to validate SAMLRequest"),
			Cause:    err,
		}
	}
	return authnRequest, relayState, nil
}
