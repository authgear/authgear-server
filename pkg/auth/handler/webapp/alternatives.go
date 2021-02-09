package webapp

import (
	"strconv"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
)

func handleAlternativeSteps(ctrl *Controller) {
	ctrl.PostAction("choose_step", func() (err error) {
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

		case webapp.SessionStepSetupOOBOTPEmail,
			webapp.SessionStepSetupOOBOTPSMS,
			webapp.SessionStepCreatePassword:
			// Simple redirect.
			choiceStep = webapp.SessionStepCreateAuthenticator
			inputFn = nil

		case webapp.SessionStepSetupTOTP:
			// Generate TOTP secret.
			choiceStep = webapp.SessionStepCreateAuthenticator
			inputFn = func() (interface{}, error) {
				return &InputSelectTOTP{}, nil
			}

		case webapp.SessionStepEnterOOBOTPAuthnEmail,
			webapp.SessionStepEnterOOBOTPAuthnSMS,
			webapp.SessionStepEnterOOBOTPSetupEmail,
			webapp.SessionStepEnterOOBOTPSetupSMS:
			// Trigger OOB-OTP code sending.
			if stepKind == webapp.SessionStepEnterOOBOTPAuthnEmail ||
				stepKind == webapp.SessionStepEnterOOBOTPAuthnSMS {
				choiceStep = webapp.SessionStepAuthenticate
			} else {
				choiceStep = webapp.SessionStepCreateAuthenticator
			}
			index, err := strconv.Atoi(ctrl.request.Form.Get("x_authenticator_index"))
			if err != nil {
				index = 0
			}
			inputFn = func() (interface{}, error) {
				return &InputTriggerOOB{
					AuthenticatorType:  ctrl.request.Form.Get("x_authenticator_type"),
					AuthenticatorIndex: index,
				}, nil
			}
		}

		// Rewind session back to the choosing step.
		originalSteps := session.Steps
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
		ctrl.skipRewind = true

		defer func() {
			// Rollback the rewound steps if processing failed.
			if e := recover(); e != nil {
				session.Steps = originalSteps
				panic(e)
			} else if err != nil {
				session.Steps = originalSteps
			}
		}()

		var result *webapp.Result
		if inputFn == nil {
			session.Steps = append(session.Steps, webapp.NewSessionStep(
				stepKind,
				session.CurrentStep().GraphID,
			))
			if err = ctrl.Page.UpdateSession(session); err != nil {
				return err
			}
			result = &webapp.Result{
				RedirectURI:      session.CurrentStep().URL().String(),
				NavigationAction: "replace",
			}
		} else {
			result, err = ctrl.InteractionPost(inputFn)
			if err != nil {
				return err
			}
			result.NavigationAction = "replace"
		}

		result.WriteResponse(ctrl.response, ctrl.request)
		return nil
	})
}
