package api

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureMagicLinkVerificationRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST").
		WithPathPattern("/api/login_link/verification")
}

var MagicLinkVerificationAPIRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"code": { "type": "string" }
		},
		"required": ["code"]
	}
`)

type MagicLinkVerificationRequest struct {
	Code string `json:"code"`
}

type MagicLinkVerificationAPIHandlerLogger struct{ *log.Logger }

func NewMagicLinkVerificationAPILogger(lf *log.Factory) MagicLinkVerificationAPIHandlerLogger {
	return MagicLinkVerificationAPIHandlerLogger{lf.New("handler-magic-link-verification")}
}

type MagicLinkVerificationAPIHandler struct {
	Logger                      MagicLinkVerificationAPIHandlerLogger
	MagicLinkOTPCodeService     otp.Service
	GlobalSessionServiceFactory *webapp.GlobalSessionServiceFactory
	JSON                        JSONResponseWriter
}

func (h *MagicLinkVerificationAPIHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var payload MagicLinkVerificationRequest
	err := httputil.BindJSONBody(req, resp, MagicLinkVerificationAPIRequestSchema.Validator(), &payload)
	if err != nil {
		if !apierrors.IsAPIError(err) {
			h.Logger.WithError(err).Error("magic link verification handler failed")
		}
		h.JSON.WriteResponse(resp, &api.Response{Error: err})
		return
	}

	_, err = h.MagicLinkOTPCodeService.SetUserInputtedMagicLinkCode(payload.Code)
	if err != nil {
		if !apierrors.IsAPIError(err) {
			h.Logger.WithError(err).Error("magic link verification handler failed")
		}
		h.JSON.WriteResponse(resp, &api.Response{Error: err})
	}

	codeModel, err := h.MagicLinkOTPCodeService.VerifyMagicLinkCode(payload.Code, true)
	if err != nil {
		if !apierrors.IsAPIError(err) {
			h.Logger.WithError(err).Error("magic link verification handler failed")
		}
		h.JSON.WriteResponse(resp, &api.Response{Error: err})
		return
	}

	// Update the web session and trigger the refresh event
	webSessionProvider := h.GlobalSessionServiceFactory.NewGlobalSessionService(
		config.AppID(codeModel.AppID),
	)
	webSession, err := webSessionProvider.GetSession(codeModel.WebSessionID)
	if err != nil {
		h.JSON.WriteResponse(resp, &api.Response{Error: err})
		return
	}
	err = webSessionProvider.UpdateSession(webSession)
	if err != nil {
		h.JSON.WriteResponse(resp, &api.Response{Error: err})
		return
	}
}
