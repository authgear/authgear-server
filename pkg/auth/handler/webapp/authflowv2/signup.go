package authflowv2

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type AuthflowV2SignupHandler struct {
	SignupLoginHandler   InternalAuthflowV2SignupLoginHandler
	AuthenticationConfig *config.AuthenticationConfig
	UIConfig             *config.UIConfig
}

func (h *AuthflowV2SignupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.AuthenticationConfig.PublicSignupDisabled {
		path := "/login"
		u := webapp.MakeRelativeURL(path, webapp.PreserveQuery(r.URL.Query()))
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
