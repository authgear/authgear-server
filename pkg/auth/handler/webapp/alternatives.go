package webapp

import (
	"net/http"
	"strconv"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
)

func handleAlternativeSteps(ctrl *Controller) {
	ctrl.PostAction("choose_step", func() error {
		session, err := ctrl.InteractionSession()
		if err != nil {
			return err
		}

		stepKind := webapp.SessionStepKind(ctrl.request.Form.Get("x_step"))
		var choiceStep webapp.SessionStepKind
		var inputFn func() (interface{}, error)
		switch stepKind {
		case webapp.SessionStepEnterTOTP,
			webapp.SessionStepEnterPassword,
			webapp.SessionStepEnterRecoveryCode:
			// Simple redirect.
			choiceStep = webapp.SessionStepAuthenticate
			inputFn = nil

		case webapp.SessionStepSetupOOBOTP,
			webapp.SessionStepCreatePassword:
			// Simple redirect.
			choiceStep = webapp.SessionStepCreateAuthenticator
			inputFn = nil

		case webapp.SessionStepSetupTOTP:
			// Generate TOTP secret.
			choiceStep = webapp.SessionStepAuthenticate
			inputFn = func() (interface{}, error) {
				return &InputSelectTOTP{}, nil
			}

		case webapp.SessionStepEnterOOBOTP:
			// Trigger OOB-OTP code sending.
			choiceStep = webapp.SessionStepAuthenticate
			index, err := strconv.Atoi(ctrl.request.Form.Get("x_authenticator_index"))
			if err != nil {
				index = 0
			}
			inputFn = func() (interface{}, error) {
				return &InputTriggerOOB{AuthenticatorIndex: index}, nil
			}
		}

		// Rewind session back to the choosing step.
		rewound := false
		for i := len(session.Steps) - 1; i >= 0; i-- {
			if session.Steps[i].Kind == choiceStep {
				session.Steps = session.Steps[:i+1]
				rewound = true
				break
			}
		}
		if !rewound {
			return webapp.ErrSessionStepMismatch
		}

		if inputFn == nil {
			session.Steps = append(session.Steps, webapp.SessionStep{
				Kind:    stepKind,
				GraphID: session.CurrentStep().GraphID,
			})
			http.Redirect(
				ctrl.response,
				ctrl.request,
				session.CurrentStep().URL().String(),
				http.StatusFound,
			)
			return ctrl.Page.UpdateSession(session)
		}

		result, err := ctrl.InteractionPost(inputFn)
		if err != nil {
			return err
		}

		result.WriteResponse(ctrl.response, ctrl.request)
		return nil
	})
}
