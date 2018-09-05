package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/skygeario/skygear-server/pkg/server/skydb"

	"github.com/skygeario/skygear-server/pkg/auth/provider"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/server/authtoken"
)

func AttachMeHandler(
	server *server.Server,
	authDependency provider.AuthProviders,
) *server.Server {
	server.Handle("/me", &MeHandlerFactory{
		authDependency,
	}).Methods("POST")
	return server
}

type MeHandlerFactory struct {
	Dependency provider.AuthProviders
}

func (f MeHandlerFactory) NewHandler(ctx context.Context, tenantConfig config.TenantConfiguration) handler.Handler {
	h := &MeHandler{}
	handler.DefaultInject(h, f.Dependency, ctx, tenantConfig)
	return h
}

// MeHandler handles me request
type MeHandler struct {
	auth.TokenStore    `dependency:"TokenStore"`
	auth.AuthInfoStore `dependency:"AuthInfoStore"`
}

func (h MeHandler) Handle(ctx handler.Context) {
	input, _ := ioutil.ReadAll(ctx.Request.Body)

	token := authtoken.Token{}

	err := h.TokenStore.Get(string(input), &token)
	if err != nil {
		// TODO:
		// handle error properly
		panic(err)
	}

	authInfo := skydb.AuthInfo{}
	err = h.AuthInfoStore.GetAuth(token.AuthInfoID, &authInfo)
	if err != nil {
		// TODO:
		// handle error properly
		panic(err)
	}

	output, err := json.Marshal(authInfo)
	if err != nil {
		// TODO:
		// handle error properly
		panic(err)
	}

	fmt.Fprint(ctx.ResponseWriter, string(output))
}
