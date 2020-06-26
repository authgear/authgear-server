package oauth

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/challenge"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
	"github.com/skygeario/skygear-server/pkg/httproute"
)

func ConfigureChallengeRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST").
		WithPathPattern("/oauth2/challenge")
}

type ChallengeRequest struct {
	Purpose challenge.Purpose `json:"purpose"`
}

func (p *ChallengeRequest) Validate() []validation.ErrorCause {
	if !p.Purpose.IsValid() {
		return []validation.ErrorCause{{
			Kind:    validation.ErrorGeneral,
			Pointer: "/purpose",
			Message: "unknown challenge purpose",
		}}
	}
	return nil
}

var challengeValidator = validation.NewValidator("http://v2.skgyear.io")

func init() {
	challengeValidator.AddSchemaFragments(
		ChallengeRequestSchema,
	)
}

// @JSONSchema
const ChallengeRequestSchema = `
{
	"$id": "#OAuthChallengeRequest",
	"type": "object",
	"properties": {
		"purpose": { "type": "string" }
	},
	"required": ["purpose"]
}
`

type ChallengeResponse struct {
	Token    string    `json:"token"`
	ExpireAt time.Time `json:"expire_at"`
}

// @JSONSchema
const ChallengeResponseSchema = `
{
	"$id": "#OAuthChallengeResponse",
	"type": "object",
	"properties": {
		"token": { "type": "string" },
		"expire_at": { "type": "string" }
	},
	"required": ["token", "expire_at"]
}
`

type ChallengeProvider interface {
	Create(purpose challenge.Purpose) (*challenge.Challenge, error)
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
	Challenges ChallengeProvider
}

func (h *ChallengeHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	result, err := h.Handle(resp, req)
	if err == nil {
		handler.WriteResponse(resp, handler.APIResponse{Result: result})
	} else {
		handler.WriteResponse(resp, handler.APIResponse{Error: err})
	}
}

func (h *ChallengeHandler) Handle(resp http.ResponseWriter, req *http.Request) (*ChallengeResponse, error) {
	var payload ChallengeRequest
	if err := handler.BindJSONBody(req, resp, challengeValidator, "#OAuthChallengeRequest", &payload); err != nil {
		return nil, err
	}

	c, err := h.Challenges.Create(payload.Purpose)
	if err != nil {
		return nil, err
	}

	return &ChallengeResponse{Token: c.Token, ExpireAt: c.ExpireAt}, nil
}
