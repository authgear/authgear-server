package oauth

import (
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	tenantdb "github.com/authgear/authgear-server/pkg/lib/infra/db/tenant"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureChallengeRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST").
		WithPathPattern("/oauth2/challenge")
}

var ChallengeAPIRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"purpose": { "type": "string" }
		},
		"required": ["purpose"]
	}
`)

var ChallengeAPIResponseSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"token": { "type": "string" },
			"expire_at": { "type": "string" }
		},
		"required": ["token", "expire_at"]
	}
`)

type ChallengeRequest struct {
	Purpose challenge.Purpose `json:"purpose"`
}

func (p *ChallengeRequest) Validate(ctx *validation.Context) {
	if !p.Purpose.IsValid() {
		ctx.Child("purpose").EmitErrorMessage("unknown challenge purpose")
	}
}

type ChallengeResponse struct {
	Token    string    `json:"token"`
	ExpireAt time.Time `json:"expire_at"`
}

type ChallengeProvider interface {
	Create(purpose challenge.Purpose) (*challenge.Challenge, error)
}

type JSONResponseWriter interface {
	WriteResponse(rw http.ResponseWriter, resp *api.Response)
}

/*
	@Operation POST /challenge - Obtain new challenge
		Obtain a new challenge for challenge-based OAuth authentication.
		Challenges can be used once only.

		@Tag User

		@RequestBody
			Describe purpose of the challenge.
			@JSONSchema {OAuthChallengeRequest}

		@Response 200
			Created challenge information.
			@JSONSchema {OAuthChallengeResponse}
*/
type ChallengeHandler struct {
	Database   *tenantdb.Handle
	Challenges ChallengeProvider
	JSON       JSONResponseWriter
}

func (h *ChallengeHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var result *ChallengeResponse
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

func (h *ChallengeHandler) Handle(resp http.ResponseWriter, req *http.Request) (*ChallengeResponse, error) {
	var payload ChallengeRequest
	if err := httputil.BindJSONBody(req, resp, ChallengeAPIRequestSchema.Validator(), &payload); err != nil {
		return nil, err
	}

	c, err := h.Challenges.Create(payload.Purpose)
	if err != nil {
		return nil, err
	}

	return &ChallengeResponse{Token: c.Token, ExpireAt: c.ExpireAt}, nil
}
