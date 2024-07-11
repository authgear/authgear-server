package oauth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureAppSessionTokenRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST").
		WithPathPattern("/oauth2/app_session_token")
}

var AppSessionTokenAPIRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"refresh_token": { "type": "string" }
		},
		"required": ["refresh_token"]
	}
`)

var AppSessionTokenAPIResponseSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"app_session_token": { "type": "string" },
			"expire_at": { "type": "string" }
		},
		"required": ["app_session_token", "expire_at"]
	}
`)

type AppSessionTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AppSessionTokenResponse struct {
	AppSessionToken string    `json:"app_session_token"`
	ExpireAt        time.Time `json:"expire_at"`
}

type AppSessionTokenIssuer interface {
	IssueAppSessionToken(ctx context.Context, refreshToken string) (string, *oauth.AppSessionToken, error)
}

type AppSessionTokenHandler struct {
	Database         *appdb.Handle
	JSON             JSONResponseWriter
	AppSessionTokens AppSessionTokenIssuer
}

func (h *AppSessionTokenHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var result *AppSessionTokenResponse
	err := h.Database.WithTx(func() (err error) {
		result, err = h.Handle(resp, req)
		return err
	})
	if err == nil {
		h.JSON.WriteResponse(resp, &api.Response{Result: result})
	} else {
		h.JSON.WriteResponse(resp, &api.Response{Error: err})
	}
}

func (h *AppSessionTokenHandler) Handle(resp http.ResponseWriter, req *http.Request) (*AppSessionTokenResponse, error) {
	var payload AppSessionTokenRequest
	if err := httputil.BindJSONBody(req, resp, AppSessionTokenAPIRequestSchema.Validator(), &payload); err != nil {
		return nil, err
	}

	token, sToken, err := h.AppSessionTokens.IssueAppSessionToken(req.Context(), payload.RefreshToken)
	var oauthError *protocol.OAuthProtocolError
	if errors.Is(err, handler.ErrInvalidRefreshToken) {
		return nil, InvalidGrant.New(err.Error())
	} else if errors.As(err, &oauthError) {
		return nil, apierrors.NewForbidden(oauthError.Error())
	} else if err != nil {
		return nil, err
	}

	return &AppSessionTokenResponse{
		AppSessionToken: token,
		ExpireAt:        sToken.ExpireAt,
	}, nil
}
