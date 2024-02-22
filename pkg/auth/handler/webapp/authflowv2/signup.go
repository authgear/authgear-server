package authflowv2

import (
	"net/http"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type AuthflowV2SignupHandler struct {
	SignupLoginHandler InternalAuthflowV2SignupLoginHandler
	UIConfig           *config.UIConfig
}

func (h *AuthflowV2SignupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flowType := authflow.FlowTypeSignup
	canSwitchToLogin := true
	uiVariant := AuthflowV2SignupUIVariantSignup
	if h.UIConfig.CombineSignupLoginFlow {
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
