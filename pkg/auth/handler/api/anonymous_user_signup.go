package api

import (
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
			}
		},
		"required": ["client_id", "session_type"]
	}
`)

var AnonymousUserSignupAPIResponseSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"token_type": { "type": "string" },
			"access_token": { "type": "string" },
			"refresh_token": { "type": "string" },
			"expires_in": { "type": "string" }
		},
		"required": ["token_type", "access_token", "refresh_token", "expires_in"]
	}
`)

type AnonymousUserSignupAPIRequest struct {
	ClientID    string                      `json:"client_id"`
	SessionType oauthhandler.WebSessionType `json:"session_type"`
}

func (p *AnonymousUserSignupAPIRequest) Validate(ctx *validation.Context) {
	if !p.SessionType.IsValid() {
		ctx.Child("session_type").EmitErrorMessage("unknown session type")
	}
}

type JSONResponseWriter interface {
	WriteResponse(rw http.ResponseWriter, resp *api.Response)
}

type AnonymousUserSignupAPIHandlerLogger struct{ *log.Logger }

func NewAnonymousUserSignupAPIHandler(lf *log.Factory) AnonymousUserSignupAPIHandlerLogger {
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

	var result *oauthhandler.SignupAnonymousUserResult
	err = h.Database.WithTx(func() error {
		result, err = h.AnonymousUserHandler.SignupAnonymousUser(
			req,
			payload.ClientID,
			payload.SessionType,
			"",
		)
		return err
	})

	if err == nil {
		h.JSON.WriteResponse(resp, &api.Response{Result: result.TokenResponse})
	} else {
		if !apierrors.IsAPIError(err) {
			h.Logger.WithError(err).Error("anonymous user signup handler failed")
		}
		h.JSON.WriteResponse(resp, &api.Response{Error: err})
	}
}
