package oauth

import (
	"errors"
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
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

const (
	AppSessionTokenAPISchemaIDRequest  = "AppSessionTokenChallengeRequest"
	AppSessionTokenAPISchemaIDResponse = "AppSessionTokenChallengeResponse"
)

var AppSessionTokenAPISchema = validation.NewMultipartSchema("").
	Add(AppSessionTokenAPISchemaIDRequest, `
		{
			"type": "object",
			"additionalProperties": false,
			"properties": {
				"refresh_token": { "type": "string" }
			},
			"required": ["refresh_token"]
		}
	`).
	Add(AppSessionTokenAPISchemaIDResponse, `
		{
			"type": "object",
			"properties": {
				"app_session_token": { "type": "string" },
				"expire_at": { "type": "string" }
			},
			"required": ["app_session_token", "expire_at"]
		}
	`).
	Instantiate()

type AppSessionTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AppSessionTokenResponse struct {
	AppSessionToken string    `json:"app_session_token"`
	ExpireAt        time.Time `json:"expire_at"`
}

type AppSessionTokenIssuer interface {
	IssueAppSessionToken(refreshToken string) (string, *oauth.AppSessionToken, error)
}

type AppSessionTokenHandler struct {
	JSON             JSONResponseWriter
	AppSessionTokens AppSessionTokenIssuer
}

func (h *AppSessionTokenHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	result, err := h.Handle(resp, req)
	if err == nil {
		h.JSON.WriteResponse(resp, &api.Response{Result: result})
	} else {
		h.JSON.WriteResponse(resp, &api.Response{Error: err})
	}
}

func (h *AppSessionTokenHandler) Handle(resp http.ResponseWriter, req *http.Request) (*AppSessionTokenResponse, error) {
	var payload AppSessionTokenRequest
	if err := httputil.BindJSONBody(req, resp, AppSessionTokenAPISchema.PartValidator(AppSessionTokenAPISchemaIDRequest), &payload); err != nil {
		return nil, err
	}

	token, sToken, err := h.AppSessionTokens.IssueAppSessionToken(payload.RefreshToken)
	var oauthError *protocol.OAuthProtocolError
	if errors.As(err, &oauthError) {
		return nil, apierrors.NewForbidden(oauthError.Error())
	} else if err != nil {
		return nil, err
	}

	return &AppSessionTokenResponse{
		AppSessionToken: token,
		ExpireAt:        sToken.ExpireAt,
	}, nil
}
