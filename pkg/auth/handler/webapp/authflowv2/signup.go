package authflowv2

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureAuthflowV2SignupRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST", "GET").WithPathPattern(AuthflowV2RouteSignup)
}

type AuthflowV2SignupHandler struct {
	SignupLoginHandler   InternalAuthflowV2SignupLoginHandler
	AuthenticationConfig *config.AuthenticationConfig
	UIConfig             *config.UIConfig
}

func (h *AuthflowV2SignupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Signup is not available when the authorization pinned a specific user via
	// id_token_hint (UserIDHint): the only valid outcome is authenticating as
	// that user. The login page already hides the signup link in this case
	// (AllowLoginOnly); this also covers direct navigation and stale links, and
	// avoids starting a signup flow that IntentSignupFlow would immediately
	// reject.
	s := webapp.GetSession(r.Context())
	userIDHintPresent := s != nil && s.UserIDHint != ""

	if h.AuthenticationConfig.PublicSignupDisabled || userIDHintPresent {
		path := "/login"
		u := webapp.MakeRelativeURL(path, webapp.PreserveQuery(r.URL.Query()))
		// #nosec G710 -- webapp.MakeRelativeURL only ever sets Path and RawQuery, never Scheme/Host, so u is always relative to the current origin.
		http.Redirect(w, r, u.String(), http.StatusFound)
		return
	}

	flowType := authflow.FlowTypeSignup
	canSwitchToLogin := true
	uiVariant := AuthflowV2SignupUIVariantSignup

	if h.UIConfig.SignupLoginFlowEnabled {
		flowType = authflow.FlowTypeSignupLogin
		canSwitchToLogin = false
		uiVariant = AuthflowV2SignupUIVariantSignupLogin
	}

	h.SignupLoginHandler.ServeHTTP(w, r, AuthflowV2SignupServeOptions{
		FlowType:         flowType,
		CanSwitchToLogin: canSwitchToLogin,
		UIVariant:        uiVariant,
	})
}
