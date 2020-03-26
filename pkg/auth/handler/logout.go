package handler

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

// AttachLogoutHandler attach logout handler to server
func AttachLogoutHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/logout").
		Handler(pkg.MakeHandler(authDependency, newLogoutHandler)).
		Methods("OPTIONS", "POST")
}

type logoutSessionManager interface {
	Logout(auth.AuthSession, http.ResponseWriter) error
}

/*
	@Operation POST /logout - Logout current session
		Logout current session.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@Response 200 {EmptyResponse}

		@Callback session_delete {SessionDeleteEvent}
		@Callback user_sync {UserSyncEvent}
*/
type LogoutHandler struct {
	SessionManager logoutSessionManager
	TxContext      db.TxContext
}

// ProvideAuthzPolicy provides authorization policy of handler
func (h LogoutHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.RequireValidUser
}

// DecodeRequest decode request payload
func (h LogoutHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (handler.RequestPayload, error) {
	payload := handler.EmptyRequestPayload{}
	err := handler.DecodeJSONBody(request, resp, &payload)
	return payload, err
}

func (h LogoutHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	err := db.WithTx(h.TxContext, func() error {
		return h.Handle(r, rw)
	})
	if err == nil {
		handler.WriteResponse(rw, handler.APIResponse{Result: struct{}{}})
	} else {
		handler.WriteResponse(rw, handler.APIResponse{Error: err})
	}
}

// Handle api request
func (h LogoutHandler) Handle(r *http.Request, rw http.ResponseWriter) error {
	sess := auth.GetSession(r.Context())

	return h.SessionManager.Logout(sess, rw)
}
