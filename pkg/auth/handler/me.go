package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachMeHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/me", &MeHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type MeHandlerFactory struct {
	Dependency auth.DependencyMap
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
	AuthContext          coreAuth.ContextGetter     `dependency:"AuthContextGetter"`
	RequireAuthz         handler.RequireAuthz       `dependency:"RequireAuthz"`
	TxContext            db.TxContext               `dependency:"TxContext"`
	AuthInfoStore        authinfo.Store             `dependency:"AuthInfoStore"`
	UserProfileStore     userprofile.Store          `dependency:"UserProfileStore"`
	PasswordAuthProvider password.Provider          `dependency:"PasswordAuthProvider"`
	IdentityProvider     principal.IdentityProvider `dependency:"IdentityProvider"`
}

func (h MeHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		policy.RequireValidUser,
	)
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
		authInfo, _ := h.AuthContext.AuthInfo()
		sess, _ := h.AuthContext.Session()
		principalID := sess.PrincipalID

		// Get Profile
		var userProfile userprofile.UserProfile
		if userProfile, err = h.UserProfileStore.GetUserProfile(authInfo.ID); err != nil {
			return err
		}

		var principal principal.Principal
		if principal, err = h.IdentityProvider.GetPrincipalByID(principalID); err != nil {
			return err
		}

		identity := model.NewIdentity(h.IdentityProvider, principal)
		user := model.NewUser(*authInfo, userProfile)

		resp = model.NewAuthResponseWithUserIdentity(user, identity)
		return nil
	})
	return
}
