package api

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	oauthhandler "github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureAnonymousUserSignupRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/api/anonymous_user/signup")
}

type AnonymousUserHandler interface {
	SignupAnonymousUser(
		ctx context.Context,
		req *http.Request,
		clientID string,
		sessionType oauthhandler.WebSessionType,
		refreshToken string,
	) (*oauthhandler.SignupAnonymousUserResult, error)
}

var AnonymousUserSignupAPIRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"client_id": { "type": "string" },
			"session_type": {
				"type": "string",
				"enum": ["cookie", "refresh_token"]
			},
			"refresh_token": { "type": "string" }
		},
		"required": ["session_type"],
		"allOf": [
			{
				"if": { "properties": { "session_type": { "const": "refresh_token" } } },
				"then": {
					"required": ["client_id"]
				}
			}
		]
	}
`)

var AnonymousUserSignupAPIResponseSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"token_type": { "type": "string" },
			"access_token": { "type": "string" },
			"refresh_token": { "type": "string" },
			"expires_in": { "type": "integer" }
		},
		"required": ["token_type", "access_token", "expires_in"]
	}
`)

type AnonymousUserSignupAPIRequest struct {
	ClientID     string                      `json:"client_id"`
	SessionType  oauthhandler.WebSessionType `json:"session_type"`
	RefreshToken string                      `json:"refresh_token"`
}

var AnonymousUserSignupAPIHandlerLogger = slogutil.NewLogger("handler-anonymous-user-signup")

type AnonymousUserSignupAPIHandlerAccessTokenEncoding interface {
	MakeUserAccessTokenFromPreparationResult(
		ctx context.Context,
		options oauth.MakeUserAccessTokenFromPreparationOptions,
	) (*oauth.IssueAccessGrantResult, error)
}

type AnonymousUserSignupAPIHandler struct {
	Database             *appdb.Handle
	AnonymousUserHandler AnonymousUserHandler
	AccessTokenEncoding  AnonymousUserSignupAPIHandlerAccessTokenEncoding
}

func (h *AnonymousUserSignupAPIHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := AnonymousUserSignupAPIHandlerLogger.GetLogger(ctx)
	var payload AnonymousUserSignupAPIRequest

	handleError := func(err error) {
		if !apierrors.IsAPIError(err) {
			logger.WithError(err).Error(ctx, "anonymous user signup handler failed")
		}
		httputil.WriteJSONResponse(ctx, resp, &api.Response{Error: err})
	}

	err := httputil.BindJSONBody(req, resp, AnonymousUserSignupAPIRequestSchema.Validator(), &payload)
	if err != nil {
		handleError(err)
		return
	}

	var result *oauthhandler.SignupAnonymousUserResult
	err = h.Database.WithTx(ctx, func(ctx context.Context) error {
		result, err = h.AnonymousUserHandler.SignupAnonymousUser(
			ctx,
			req,
			payload.ClientID,
			payload.SessionType,
			payload.RefreshToken,
		)
		return err
	})
	if err != nil {
		handleError(err)
		return
	}

	if result.Cookies != nil {
		// cookie
		for _, cookie := range result.Cookies {
			httputil.UpdateCookie(resp, cookie)
		}
		httputil.WriteJSONResponse(ctx, resp, &api.Response{Result: struct{}{}})
	} else {
		// refresh token

		if result.PrepareUserAccessGrantByRefreshTokenResult != nil {
			result.PrepareUserAccessGrantByRefreshTokenResult.RotateRefreshTokenResult.WriteTo(result.Response)

			a, err := h.AccessTokenEncoding.MakeUserAccessTokenFromPreparationResult(ctx, oauth.MakeUserAccessTokenFromPreparationOptions{
				ClientConfig:      result.Client,
				PreparationResult: result.PrepareUserAccessGrantByRefreshTokenResult.PreparationResult,
			})
			if err != nil {
				handleError(err)
				return
			}

			a.WriteTo(result.Response)
		}

		httputil.WriteJSONResponse(ctx, resp, &api.Response{Result: result.Response})
	}
}
