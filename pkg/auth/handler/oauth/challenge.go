package oauth

import (
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/auth/dependency/challenge"
	"github.com/authgear/authgear-server/pkg/core/handler"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureChallengeRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST").
		WithPathPattern("/oauth2/challenge")
}

const (
	ChallengeAPISchemaIDRequest  = "OAuthChallengeRequest"
	ChallengeAPISchemaIDResponse = "OAuthChallengeResponse"
)

var ChallengeAPISchema = validation.NewMultipartSchema("").
	Add(ChallengeAPISchemaIDRequest, `
		{
			"type": "object",
			"additionalProperties": false,
			"properties": {
				"purpose": { "type": "string" }
			},
			"required": ["purpose"]
		}
	`).
	Add(ChallengeAPISchemaIDResponse, `
		{
			"type": "object",
			"properties": {
				"token": { "type": "string" },
				"expire_at": { "type": "string" }
			},
			"required": ["token", "expire_at"]
		}
	`).
	Instantiate()

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
	if err := handler.BindJSONBody(req, resp, ChallengeAPISchema.PartValidator(ChallengeAPISchemaIDRequest), &payload); err != nil {
		return nil, err
	}

	c, err := h.Challenges.Create(payload.Purpose)
	if err != nil {
		return nil, err
	}

	return &ChallengeResponse{Token: c.Token, ExpireAt: c.ExpireAt}, nil
}
