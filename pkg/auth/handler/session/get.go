package session

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachGetHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/session/get").
		Handler(server.FactoryToHandler(&GetHandlerFactory{
			authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type GetHandlerFactory struct {
	Dependency pkg.DependencyMap
}

func (f GetHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &GetHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

type GetRequestPayload struct {
	SessionID string `json:"session_id"`
}

// @JSONSchema
const GetRequestSchema = `
{
	"$id": "#SessionGetRequest",
	"type": "object",
	"properties": {
		"session_id": { "type": "string", "minLength": 1 }
	},
	"required": ["session_id"]
}
`

type GetResponse struct {
	Session model.Session `json:"session"`
}

// @JSONSchema
const GetResponseSchema = `
{
	"$id": "#SessionGetResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"session": { "$ref": "#Session" }
			}
		}
	}
}
`

/*
	@Operation POST /session/get - Get current user sessions
		Get the sessions with specified ID of current user.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			Describe the session ID.
			@JSONSchema {SessionGetRequest}

		@Response 200
			The requested session.
			@JSONSchema {SessionGetResponse}
*/
type GetHandler struct {
	Validator       *validation.Validator `dependency:"Validator"`
	RequireAuthz    handler.RequireAuthz  `dependency:"RequireAuthz"`
	TxContext       db.TxContext          `dependency:"TxContext"`
	SessionProvider session.Provider      `dependency:"SessionProvider"`
}

func (h GetHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.RequireValidUser
}

func (h GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	var payload GetRequestPayload
	if err := handler.BindJSONBody(r, w, h.Validator, "#SessionGetRequest", &payload); err != nil {
		response.Error = err
	} else {
		result, err := h.Handle(r, payload)
		if err != nil {
			response.Error = err
		} else {
			response.Result = result
		}
	}
	handler.WriteResponse(w, response)
}

func (h GetHandler) Handle(r *http.Request, payload GetRequestPayload) (resp interface{}, err error) {
	err = db.WithTx(h.TxContext, func() error {
		userID := auth.GetSession(r.Context()).AuthnAttrs().UserID
		sessionID := payload.SessionID

		s, err := h.SessionProvider.Get(sessionID)
		if err != nil {
			if errors.Is(err, session.ErrSessionNotFound) {
				err = errSessionNotFound
			}
			return err
		}
		if s.UserID != userID {
			return errSessionNotFound
		}

		// TODO(authn): use new session
		// resp = GetResponse{Session: authSession.Format(s)}
		return nil
	})
	return
}
