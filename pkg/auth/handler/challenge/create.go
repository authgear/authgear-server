package challenge

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/challenge"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachCreateHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/challenge").
		Handler(auth.MakeHandler(authDependency, newCreateHandler)).
		Methods("OPTIONS", "POST")
}

type CreateRequest struct {
	Purpose challenge.Purpose `json:"purpose"`
}

func (p *CreateRequest) Validate() []validation.ErrorCause {
	if !p.Purpose.IsValid() {
		return []validation.ErrorCause{{
			Kind:    validation.ErrorGeneral,
			Pointer: "/purpose",
			Message: "unknown challenge purpose",
		}}
	}
	return nil
}

// @JSONSchema
const CreateRequestSchema = `
{
	"$id": "#CreateChallengeRequest",
	"type": "object",
	"properties": {
		"purpose": { "type": "string" }
	},
	"required": ["purpose"]
}
`

type CreateResponse struct {
	Token    string    `json:"token"`
	ExpireAt time.Time `json:"expire_at"`
}

// @JSONSchema
const CreateResponseSchema = `
{
	"$id": "#CreateChallengeResponse",
	"type": "object",
	"properties": {
		"token": { "type": "string" }
		"expire_at": { "type": "string" }
	},
	"required": ["token", "expire_at"]
}
`

type challengeProvider interface {
	Create(purpose challenge.Purpose) (*challenge.Challenge, error)
}

/*
	@Operation POST /challenge - Obtain new challenge
		Obtain a new challenge for challenge-based authentication.
		Challenges can be used once only.

		@Tag User

		@RequestBody
			Describe purpose of the challenge.
			@JSONSchema {CreateChallengeRequest}

		@Response 200
			Created challenge information.
			@JSONSchema {CreateChallengeResponse}
*/
type CreateHandler struct {
	Validator  *validation.Validator
	Challenges challengeProvider
}

func (h *CreateHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	result, err := h.Handle(resp, req)
	if err == nil {
		handler.WriteResponse(resp, handler.APIResponse{Result: result})
	} else {
		handler.WriteResponse(resp, handler.APIResponse{Error: err})
	}
}

func (h *CreateHandler) Handle(resp http.ResponseWriter, req *http.Request) (*CreateResponse, error) {
	var payload CreateRequest
	if err := handler.BindJSONBody(req, resp, h.Validator, "#CreateChallengeRequest", &payload); err != nil {
		return nil, err
	}

	c, err := h.Challenges.Create(payload.Purpose)
	if err != nil {
		return nil, err
	}

	return &CreateResponse{Token: c.Token, ExpireAt: c.ExpireAt}, nil
}
