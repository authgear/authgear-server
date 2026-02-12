package webapp

import (
	"context"
	"fmt"
	"net/http"

	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/webappoauth"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureSSOCallbackRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/sso/oauth2/callback/:alias")
}

type SSOCallbackHandlerOAuthStateStore interface {
	PopAndRecoverState(ctx context.Context, stateToken string) (state *webappoauth.WebappOAuthState, err error)
}

type SSOCallbackHandler struct {
	AuthflowController *AuthflowController
	// TODO(tung)
	ControllerFactory ControllerFactory
	ErrorRenderer     *ErrorRenderer
	OAuthStateStore   SSOCallbackHandlerOAuthStateStore
	AccountManagement *accountmanagement.Service
}

func (h *SSOCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stateToken := r.FormValue("state")
	state, err := h.OAuthStateStore.PopAndRecoverState(r.Context(), stateToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s := session.GetSession(r.Context())

	if state.AccountManagementToken != "" {
		redirectURL, err := url.Parse("/settings/identity/oauth")
		if err != nil {
			panic(err)
		}

		_, err = h.AccountManagement.FinishAddingIdentityOAuth(r.Context(), s, &accountmanagement.FinishAddingIdentityOAuthInput{
			Token: state.AccountManagementToken,
			Query: r.URL.Query().Encode(),
		})
		if err != nil {
			h.ErrorRenderer.MakeAuthflowErrorResult(r.Context(), w, r, *redirectURL, err).WriteResponse(w, r)
			return
		}

		http.Redirect(w, r, redirectURL.String(), http.StatusFound)
		return
	}

	switch state.UIImplementation {
	case config.UIImplementationAuthflowV2:
		// authflow
		h.AuthflowController.HandleOAuthCallback(r.Context(), w, r, AuthflowOAuthCallbackResponse{
			Query: r.Form.Encode(),
			State: state,
		})
	default:
		panic(fmt.Errorf("expected ui implementation to be set in state"))
	}
}
