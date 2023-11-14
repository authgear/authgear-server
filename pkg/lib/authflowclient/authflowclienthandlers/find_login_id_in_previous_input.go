package authflowclienthandlers

import (
	"github.com/authgear/authgear-server/pkg/auth/webapp"
)

func findLoginIDInPreviousInput(s *webapp.Session, xStep string) (string, bool) {
	if s.Authflow == nil {
		return "", false
	}

	for {
		screen := s.Authflow.AllScreens[xStep]
		if screen == nil {
			return "", false
		}

		if screen.PreviousInput != nil {
			previousInput := screen.PreviousInput
			if loginID, ok := previousInput["login_id"].(string); ok {
				return loginID, true
			}
		}

		if screen.BranchStateToken != nil {
			branchXStep := screen.BranchStateToken.XStep
			branchScreen := s.Authflow.AllScreens[branchXStep]
			if branchScreen != nil {
				previousInput := branchScreen.PreviousInput
				if loginID, ok := previousInput["login_id"].(string); ok {
					return loginID, true
				}
			}
		}

		// Otherwise update xStep and find recursively.
		xStep = screen.PreviousXStep
	}
}
