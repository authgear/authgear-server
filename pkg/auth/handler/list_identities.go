package handler

import (
	"net/http"
	"sort"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

func AttachListIdentitiesHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/identity/list").
		Handler(pkg.MakeHandler(authDependency, newListIdentitiesHandler)).
		Methods("OPTIONS", "POST")
}

type ListIdentityProvider interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

// @JSONSchema
const IdentityListResponseSchema = `
{
	"$id": "#IdentityListResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"identities": { 
					"type": "array",
					"items": { "$ref": "#Identity" }
				}
			}
		}
	}
}
`

type IdentityListResponse struct {
	Identities []model.Identity `json:"identities"`
}

/*
	@Operation POST /identity/list - List identities
		Returns list of identities of current user.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@Response 200
			Current user and identity info.
			@JSONSchema {IdentityListResponse}
*/
type ListIdentitiesHandler struct {
	TxContext        db.TxContext
	IdentityProvider ListIdentityProvider
}

func (h ListIdentitiesHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireValidUser
}

func (h ListIdentitiesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := h.Handle(w, r)
	if err == nil {
		handler.WriteResponse(w, handler.APIResponse{Result: result})
	} else {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
	}
}

func (h ListIdentitiesHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	if err = handler.DecodeJSONBody(r, w, &struct{}{}); err != nil {
		return
	}

	err = db.WithTx(h.TxContext, func() error {
		userID := auth.GetSession(r.Context()).AuthnAttrs().UserID

		iis, err := h.IdentityProvider.ListByUser(userID)
		if err != nil {
			return err
		}

		sort.Slice(iis, func(i, j int) bool {
			return iis[i].ID < iis[j].ID
		})

		identities := make([]model.Identity, len(iis))
		for i, ii := range iis {
			identities[i] = model.Identity{
				Type:   string(ii.Type),
				Claims: ii.Claims,
			}
		}

		resp = IdentityListResponse{Identities: identities}
		return nil
	})
	return
}
