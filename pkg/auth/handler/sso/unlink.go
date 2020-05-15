package sso

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func AttachUnlinkHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/sso/{provider}/unlink").
		Handler(pkg.MakeHandler(authDependency, newUnlinkHandler)).
		Methods("OPTIONS", "POST")
}

type OAuthUnlinkInteractionFlow interface {
	UnlinkkWithOAuthProvider(
		clientID string, userID string, oauthProviderInfo config.OAuthProviderConfiguration,
	) error
}

/*
	@Operation POST /sso/{provider_id}/unlink - Unlink SSO provider
		Unlink the specified SSO provider from the current user.

		@Tag SSO
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@Parameter {SSOProviderID}
		@Response 200 {EmptyResponse}

		@Callback identity_delete {UserSyncEvent}
		@Callback user_sync {UserSyncEvent}
*/
type UnlinkHandler struct {
	TxContext       db.TxContext
	ProviderFactory *sso.OAuthProviderFactory
	Interactions    OAuthUnlinkInteractionFlow
}

func (h UnlinkHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireValidUser
}

func (h UnlinkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	var payload struct{}
	if err := handler.DecodeJSONBody(r, w, &payload); err != nil {
		response.Error = err
	} else {
		result, err := h.Handle(r)
		if err != nil {
			response.Error = err
		} else {
			response.Result = result
		}
	}
	handler.WriteResponse(w, response)
}

func (h UnlinkHandler) Handle(r *http.Request) (resp interface{}, err error) {
	err = db.WithTx(h.TxContext, func() error {
		vars := mux.Vars(r)
		providerID := vars["provider"]

		providerConfig, ok := h.ProviderFactory.GetOAuthProviderConfig(providerID)
		if !ok {
			return skyerr.NewNotFound("unknown SSO provider")
		}

		sess := auth.GetSession(r.Context())
		userID := sess.AuthnAttrs().UserID

		err := h.Interactions.UnlinkkWithOAuthProvider(
			sess.GetClientID(), userID, providerConfig,
		)

		if err != nil {
			return err
		}

		resp = struct{}{}
		return nil
	})
	return
}
