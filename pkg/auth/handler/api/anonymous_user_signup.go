package api

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	oauthhandler "github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
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

type JSONResponseWriter interface {
	WriteResponse(rw http.ResponseWriter, resp *api.Response)
}

type AnonymousUserSignupAPIHandlerLogger struct{ *log.Logger }

func NewAnonymousUserSignupAPIHandlerLogger(lf *log.Factory) AnonymousUserSignupAPIHandlerLogger {
	return AnonymousUserSignupAPIHandlerLogger{lf.New("handler-anonymous-user-signup")}
}

type AnonymousUserSignupAPIHandler struct {
	Logger               AnonymousUserSignupAPIHandlerLogger
	Database             *appdb.Handle
	JSON                 JSONResponseWriter
	AnonymousUserHandler AnonymousUserHandler
}

func (h *AnonymousUserSignupAPIHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var payload AnonymousUserSignupAPIRequest
	err := httputil.BindJSONBody(req, resp, AnonymousUserSignupAPIRequestSchema.Validator(), &payload)
	if err != nil {
		h.JSON.WriteResponse(resp, &api.Response{Error: err})
		return
	}

	ctx := req.Context()
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

	if err == nil {
		if result.Cookies != nil {
			// cookie
			for _, cookie := range result.Cookies {
				httputil.UpdateCookie(resp, cookie)
			}
			h.JSON.WriteResponse(resp, &api.Response{Result: struct{}{}})
		} else {
			// refresh token
			h.JSON.WriteResponse(resp, &api.Response{Result: result.TokenResponse})
		}
	} else {
		if !apierrors.IsAPIError(err) {
			h.Logger.WithError(err).Error("anonymous user signup handler failed")
		}
		h.JSON.WriteResponse(resp, &api.Response{Error: err})
	}
}
