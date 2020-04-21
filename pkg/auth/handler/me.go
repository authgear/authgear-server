package handler

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachMeHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/me").
		Handler(server.FactoryToHandler(&MeHandlerFactory{
			authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type MeHandlerFactory struct {
	Dependency pkg.DependencyMap
}

func (f MeHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &MeHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

/*
	@Operation POST /me - Get current user information
		Returns information on current user and identity.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@Response 200
			Current user and identity info.
			@JSONSchema {UserIdentityResponse}
*/
type MeHandler struct {
	RequireAuthz         handler.RequireAuthz       `dependency:"RequireAuthz"`
	TxContext            db.TxContext               `dependency:"TxContext"`
	AuthInfoStore        authinfo.Store             `dependency:"AuthInfoStore"`
	UserProfileStore     userprofile.Store          `dependency:"UserProfileStore"`
	PasswordAuthProvider password.Provider          `dependency:"PasswordAuthProvider"`
	IdentityProvider     principal.IdentityProvider `dependency:"IdentityProvider"`
}

func (h MeHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireValidUser
}

func (h MeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := h.Handle(w, r)
	if err == nil {
		handler.WriteResponse(w, handler.APIResponse{Result: result})
	} else {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
	}
}

func (h MeHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	if err = handler.DecodeJSONBody(r, w, &struct{}{}); err != nil {
		return
	}

	err = db.WithTx(h.TxContext, func() error {
		sess := auth.GetSession(r.Context())

		authInfo := &authinfo.AuthInfo{}
		if err := h.AuthInfoStore.GetAuth(sess.AuthnAttrs().UserID, authInfo); err != nil {
			return err
		}

		var userProfile userprofile.UserProfile
		if userProfile, err = h.UserProfileStore.GetUserProfile(sess.AuthnAttrs().UserID); err != nil {
			return err
		}

		identity := model.NewIdentityFromAttrs(sess.AuthnAttrs())
		user := model.NewUser(*authInfo, userProfile)

		resp = model.NewAuthResponseWithUserIdentity(user, identity)
		return nil
	})
	return
}
