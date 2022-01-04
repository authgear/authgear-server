package api

import (
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	oauthhandler "github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureAnonymousUserPromotionCodeRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/api/anonymous_user/promotion_code")
}

var AnonymousUserPromotionCodeAPIRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"session_type": { "type": "string" },
			"refresh_token": { "type": "string" }
		},
		"required": ["session_type"]
	}
`)

var AnonymousUserPromotionCodeResponseSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"promotion_code": { "type": "string" },
			"expire_at": { "type": "string" }
		},
		"required": ["promotion_code", "expire_at"]
	}
`)

type AnonymousUserPromotionCodeRequest struct {
	SessionType  oauthhandler.WebSessionType `json:"session_type"`
	RefreshToken string                      `json:"refresh_token"`
}

func (p *AnonymousUserPromotionCodeRequest) Validate(ctx *validation.Context) {
	if !p.SessionType.IsValid() {
		ctx.Child("session_type").EmitErrorMessage("unknown session type")
	}
}

type AnonymousUserPromotionCodeResponse struct {
	PromotionCode string    `json:"promotion_code"`
	ExpireAt      time.Time `json:"expire_at"`
}

type AnonymousUserPromotionCodeAPIHandlerLogger struct{ *log.Logger }

func NewAnonymousUserPromotionCodeAPILogger(lf *log.Factory) AnonymousUserPromotionCodeAPIHandlerLogger {
	return AnonymousUserPromotionCodeAPIHandlerLogger{lf.New("handler-anonymous-user-promotion-code")}
}

type AnonymousUserPromotionCodeAPIHandler struct {
	Logger   AnonymousUserPromotionCodeAPIHandlerLogger
	Database *appdb.Handle
	JSON     JSONResponseWriter
}

func (h *AnonymousUserPromotionCodeAPIHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var payload AnonymousUserPromotionCodeRequest
	err := httputil.BindJSONBody(req, resp, AnonymousUserPromotionCodeAPIRequestSchema.Validator(), &payload)
	if err != nil {
		h.JSON.WriteResponse(resp, &api.Response{Error: err})
		return
	}

	result := &AnonymousUserPromotionCodeResponse{}
	err = h.Database.WithTx(func() error {
		// FIXME(anonymous-user): promotion code implementation
		return nil
	})

	if err == nil {
		h.JSON.WriteResponse(resp, &api.Response{Result: result})
	} else {
		if !apierrors.IsAPIError(err) {
			h.Logger.WithError(err).Error("anonymous user promotion code handler failed")
		}
		h.JSON.WriteResponse(resp, &api.Response{Error: err})
	}
}
