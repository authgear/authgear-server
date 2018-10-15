package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func AttachMeHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/me", &MeHandlerFactory{
		authDependency,
	}).Methods("POST")
	return server
}

type MeHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f MeHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &MeHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f MeHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

// MeHandler handles me request
type MeHandler struct {
	AuthContext   coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	TxContext     db.TxContext           `dependency:"TxContext"`
	TokenStore    authtoken.Store        `dependency:"TokenStore"`
	AuthInfoStore authinfo.Store         `dependency:"AuthInfoStore"`
}

func (h MeHandler) WithTx() bool {
	return true
}

func (h MeHandler) DecodeRequest(request *http.Request) (payload handler.RequestPayload, err error) {
	payload = handler.EmptyRequestPayload{}
	return
}

func (h MeHandler) Handle(req interface{}) (resp interface{}, err error) {
	authInfo := h.AuthContext.AuthInfo()

	// refresh access token with a newly generated one
	token, err := h.TokenStore.NewToken(authInfo.ID)
	if err != nil {
		panic(err)
	}

	if err = h.TokenStore.Put(&token); err != nil {
		panic(err)
	}

	resp = response.NewAuthResponse(*authInfo, skydb.Record{}, token.AccessToken)

	// Populate the activity time to user
	now := timeNow()
	authInfo.LastSeenAt = &now
	if err := h.AuthInfoStore.UpdateAuth(authInfo); err != nil {
		err = skyerr.MakeError(err)
	}

	return
}
