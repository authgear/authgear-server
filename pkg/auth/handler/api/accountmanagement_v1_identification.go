package api

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureAccountManagementV1IdentificationRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST").WithPathPattern("/api/v1/account/identification")
}

var AccountManagementV1IdentificationSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"identification": {
				"type": "string",
				"const": "oauth"
			},
			"alias": {
				"type": "string",
				"minLength": 1
			},
			"redirect_uri": {
				"type": "string",
				"format": "uri"
			}
		},
		"required": ["identification", "alias", "redirect_uri"]
	}
`)

type AccountManagementV1IdentificationRequest struct {
	Identification string `json:"identification,omitempty"`
	Alias          string `json:"alias,omitempty"`
	RedirectURI    string `json:"redirect_uri,omitempty"`
}

type AccountManagementV1IdentificationHandler struct {
	JSON JSONResponseWriter
}

func (h *AccountManagementV1IdentificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var request AccountManagementV1IdentificationRequest
	err = httputil.BindJSONBody(r, w, AccountManagementV1IdentificationSchema.Validator(), &request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}
}
