package api

import (
	"context"
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	oauthhandler "github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
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
			"session_type": {
				"type": "string",
				"enum": ["cookie", "refresh_token"]
			},
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

type AnonymousUserPromotionCodeResponse struct {
	PromotionCode string    `json:"promotion_code"`
	ExpireAt      time.Time `json:"expire_at"`
}

type PromotionCodeIssuer interface {
	IssuePromotionCode(
		ctx context.Context,
		req *http.Request,
		sessionType oauthhandler.WebSessionType,
		refreshToken string,
	) (code string, codeObj *anonymous.PromotionCode, err error)
}

var AnonymousUserPromotionCodeAPIHandlerLogger = slogutil.NewLogger("handler-anonymous-user-promotion-code")

type AnonymousUserPromotionCodeAPIHandler struct {
	Database       *appdb.Handle
	PromotionCodes PromotionCodeIssuer
}

func (h *AnonymousUserPromotionCodeAPIHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	var payload AnonymousUserPromotionCodeRequest
	err := httputil.BindJSONBody(req, resp, AnonymousUserPromotionCodeAPIRequestSchema.Validator(), &payload)
	if err != nil {
		httputil.WriteJSONResponse(ctx, resp, &api.Response{Error: err})
		return
	}

	logger := AnonymousUserPromotionCodeAPIHandlerLogger.GetLogger(ctx)

	result := &AnonymousUserPromotionCodeResponse{}
	err = h.Database.WithTx(ctx, func(ctx context.Context) error {
		code, codeObj, err := h.PromotionCodes.IssuePromotionCode(
			ctx,
			req,
			payload.SessionType,
			payload.RefreshToken,
		)
		if err != nil {
			return err
		}
		result.PromotionCode = code
		result.ExpireAt = codeObj.ExpireAt
		return nil
	})

	if err == nil {
		httputil.WriteJSONResponse(ctx, resp, &api.Response{Result: result})
	} else {
		if !apierrors.IsAPIError(err) {
			logger.WithError(err).Error(ctx, "anonymous user promotion code handler failed")
		}
		httputil.WriteJSONResponse(ctx, resp, &api.Response{Error: err})
	}
}
