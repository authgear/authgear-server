package session

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachRevokeHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/session/revoke").
		Handler(pkg.MakeHandler(authDependency, newRevokeHandler)).
		Methods("OPTIONS", "POST")
}

type RevokeRequestPayload struct {
	CurrentSessionID string `json:"-"`
	SessionID        string `json:"session_id"`
}

func (p *RevokeRequestPayload) Validate() []validation.ErrorCause {
	if p.CurrentSessionID != p.SessionID {
		return nil
	}
	return []validation.ErrorCause{{
		Kind:    validation.ErrorGeneral,
		Pointer: "/session_id",
		Message: "session_id must not be current session",
	}}
}

// @JSONSchema
const RevokeRequestSchema = `
{
	"$id": "#SessionRevokeRequest",
	"type": "object",
	"properties": {
		"session_id": { "type": "string", "minLength": 1 }
	},
	"required": ["session_id"]
}
`

type sessionRevokeManager interface {
	Get(id string) (auth.AuthSession, error)
	Revoke(auth.AuthSession) error
}

/*
	@Operation POST /session/revoke - Revoke session
		Update specified session. Current session cannot be revoked.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			Describe the session ID.
			@JSONSchema {SessionRevokeRequest}

		@Response 200 {EmptyResponse}
*/
type RevokeHandler struct {
	validator      *validation.Validator
	txContext      db.TxContext
	sessionManager sessionRevokeManager
}

func (h *RevokeHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.RequireValidUser
}

func (h *RevokeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r.Context())

	var payload RevokeRequestPayload
	payload.CurrentSessionID = session.SessionID()
	if err := handler.BindJSONBody(r, rw, h.validator, "#SessionRevokeRequest", &payload); err != nil {
		handler.WriteResponse(rw, handler.APIResponse{Error: err})
		return
	}

	err := db.WithTx(h.txContext, func() error {
		userID := session.AuthnAttrs().UserID

		// ignore session not found errors
		session, err := h.sessionManager.Get(payload.SessionID)
		if errors.Is(err, auth.ErrSessionNotFound) {
			return nil
		} else if err != nil {
			return err
		}

		if session.AuthnAttrs().UserID != userID {
			return nil
		}

		err = h.sessionManager.Revoke(session)
		if err != nil {
			return err
		}

		return nil
	})

	if err == nil {
		handler.WriteResponse(rw, handler.APIResponse{Result: struct{}{}})
	} else {
		handler.WriteResponse(rw, handler.APIResponse{Error: err})
	}
}
