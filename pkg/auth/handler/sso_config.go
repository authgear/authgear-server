package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachSSOConfigHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.
		Handle("/sso/config", &SSOConfigHandler{}).
		Methods("OPTIONS", "POST")
	return server
}

type SSOConfigHandler struct {
}

// NewHandler returns the SSO configs.
//
// curl \
//   -X POST \
//   -H "Content-Type: application/json" \
//   -H "X-Skygear-Api-Key: API_KEY" \
//   -d @- \
//   http://localhost:3000/sso/config \
// <<EOF
// {
// }
// EOF
//
// {
//     "result": [
//         "http://localhost",
//         "http://127.0.0.1"
//     ]
// }
func (f SSOConfigHandler) NewHandler(request *http.Request) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		tConfig := config.GetTenantConfig(r)
		var response handler.APIResponse
		response.Result = tConfig.SSOSetting.AllowedCallbackURLs
		handler.WriteResponse(rw, response)
	})
}

func (f SSOConfigHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}
