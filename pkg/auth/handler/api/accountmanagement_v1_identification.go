package api

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

//go:generate mockgen -source=accountmanagement_v1_identification.go -destination=accountmanagement_v1_identification_mock_test.go -package api

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
			},
			"exclude_state_in_authorization_url": {
				"type": "boolean"
			}
		},
		"required": ["identification", "alias", "redirect_uri"]
	}
`)

type AccountManagementV1IdentificationRequest struct {
	Identification                 string `json:"identification,omitempty"`
	Alias                          string `json:"alias,omitempty"`
	RedirectURI                    string `json:"redirect_uri,omitempty"`
	ExcludeStateInAuthorizationURL *bool  `json:"exclude_state_in_authorization_url,omitempty"`
}

type AccountManagementV1IdentificationHandlerService interface {
	StartAdding(input *accountmanagement.StartAddingInput) (*accountmanagement.StartAddingOutput, error)
}

type AccountManagementV1IdentificationHandler struct {
	JSON    JSONResponseWriter
	Service AccountManagementV1IdentificationHandlerService
}

func (h *AccountManagementV1IdentificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var request AccountManagementV1IdentificationRequest
	err = httputil.BindJSONBody(r, w, AccountManagementV1IdentificationSchema.Validator(), &request)
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}
	h.handle(w, r, request)
}

func (h *AccountManagementV1IdentificationHandler) handle(w http.ResponseWriter, r *http.Request, request AccountManagementV1IdentificationRequest) {
	userID := session.GetUserID(r.Context())
	includeAndBindState := true
	if request.ExcludeStateInAuthorizationURL != nil && *request.ExcludeStateInAuthorizationURL {
		includeAndBindState = false
	}
	output, err := h.Service.StartAdding(&accountmanagement.StartAddingInput{
		UserID:      *userID,
		Alias:       request.Alias,
		RedirectURI: request.RedirectURI,
		IncludeStateAuthorizationURLAndBindStateToToken: includeAndBindState,
	})
	if err != nil {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
		return
	}

	h.JSON.WriteResponse(w, &api.Response{Result: output})
}
