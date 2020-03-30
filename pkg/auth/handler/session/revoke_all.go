package session

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

func AttachRevokeAllHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/session/revoke_all").
		Handler(pkg.MakeHandler(authDependency, newRevokeAllHandler)).
		Methods("OPTIONS", "POST")
}

type sessionRevokeAllManager interface {
	List(userID string) ([]auth.AuthSession, error)
	Revoke(auth.AuthSession) error
}

/*
	@Operation POST /session/revoke_all - Revoke all sessions
		Revoke all sessions, excluding current session.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@Response 200 {EmptyResponse}
*/
type RevokeAllHandler struct {
	txContext      db.TxContext
	sessionManager sessionRevokeAllManager
}

func (h *RevokeAllHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireValidUser
}

func (h *RevokeAllHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if err := handler.DecodeJSONBody(r, rw, &struct{}{}); err != nil {
		handler.WriteResponse(rw, handler.APIResponse{Error: err})
		return
	}

	err := db.WithTx(h.txContext, func() error {
		session := auth.GetSession(r.Context())
		userID := session.AuthnAttrs().UserID

		sessions, err := h.sessionManager.List(userID)
		if err != nil {
			return err
		}
		for _, s := range sessions {
			if s.SessionID() == session.SessionID() {
				continue
			}

			err = h.sessionManager.Revoke(s)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err == nil {
		handler.WriteResponse(rw, handler.APIResponse{Result: struct{}{}})
	} else {
		handler.WriteResponse(rw, handler.APIResponse{Error: err})
	}
}
