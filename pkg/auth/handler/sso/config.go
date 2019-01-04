package sso

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachConfigHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.
		Handle("/sso/config", &ConfigHandler{}).
		Methods("OPTIONS", "POST")
	return server
}

type ConfigHandler struct {
}

type ConfigResp struct {
	AuthorizedURLS []string `json:"authorized_urls"`
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
//     "result": {
//         "authorized_urls": [
//             "http://localhost",
//             "http://127.0.0.1"
//         }
//     }
// }
func (f ConfigHandler) NewHandler(request *http.Request) http.Handler {
	handleAPICall := func(r *http.Request) (apiResp handler.APIResponse) {
		tConfig := config.GetTenantConfig(r)
		authorizedURLs := tConfig.SSOSetting.AllowedCallbackURLs
		if authorizedURLs == nil {
			authorizedURLs = []string{}
		}
		resp := ConfigResp{
			AuthorizedURLS: authorizedURLs,
		}
		apiResp.Result = resp

		return
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		response := handleAPICall(r)
		handler.WriteResponse(rw, response)
	})
}

func (f ConfigHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf()
}
