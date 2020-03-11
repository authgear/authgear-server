package sso

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/apiclientconfig"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachConfigHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/sso/config").
		Handler(server.FactoryToHandler(&ConfigHandler{})).
		Methods("OPTIONS", "POST")
}

type ConfigHandler struct {
	ClientProvider apiclientconfig.Provider `dependency:"APIClientConfigurationProvider"`
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
		authorizedURLs := []string{}
		_, client, ok := f.ClientProvider.Get()
		if ok {
			redirectURIs := client.RedirectURIs()
			if len(redirectURIs) > 0 {
				authorizedURLs = redirectURIs
			}
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
