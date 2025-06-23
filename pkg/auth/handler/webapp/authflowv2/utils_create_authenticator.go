package authflowv2

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
)

func getTakenBranchCreateAuthenticatorAuthentication(s *webapp.AuthflowScreenWithFlowResponse) model.AuthenticationFlowAuthentication {
	// If the current step already tells the authentication, use it
	authentication := s.StateTokenFlowResponse.Action.Authentication
	if authentication == "" {
		// Else, get it from the first option of the branch step
		options := s.BranchStateTokenFlowResponse.Action.Data.(declarative.CreateAuthenticatorData).Options
		index := *s.Screen.TakenBranchIndex
		option := options[index]
		authentication = option.Authentication
	}

	return authentication
}
