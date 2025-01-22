package webapp

import (
	"context"
	"fmt"
	"strconv"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

type AuthenticationBeginNode interface {
	GetAuthenticationEdges() ([]interaction.Edge, error)
	GetAuthenticationStage() authn.AuthenticationStage
}

type CreateAuthenticatorBeginNode interface {
	GetCreateAuthenticatorEdges() ([]interaction.Edge, error)
	GetCreateAuthenticatorStage() authn.AuthenticationStage
}

type CreateAuthenticatorPhoneOTPNode interface {
	GetCreateAuthenticatorStage() authn.AuthenticationStage
	GetSelectedPhoneNumberForPhoneOTP() string
}

// nolint: gocognit
func handleAlternativeSteps(ctrl *Controller) {
	ctrl.PostAction("choose_step", func(ctx context.Context) (err error) {
		session, err := ctrl.InteractionSession(ctx)
		if err != nil {
			return err
		}

		stepKind := webapp.SessionStepKind(ctrl.request.Form.Get("x_step_kind"))
		var choiceStep webapp.SessionStepKind
		var inputFn func() (interface{}, error)
		switch stepKind {
		case webapp.SessionStepEnterTOTP,
			webapp.SessionStepEnterPassword,
			webapp.SessionStepEnterRecoveryCode:
			// Simple redirect.
			choiceStep = webapp.SessionStepAuthenticate
			inputFn = nil

		case webapp.SessionStepCreatePassword:
			// Simple redirect.
			choiceStep = webapp.SessionStepCreateAuthenticator
			inputFn = nil

		case webapp.SessionStepSetupOOBOTPEmail,
			webapp.SessionStepSetupOOBOTPSMS:
			graph, err := ctrl.InteractionGet(ctx)
			if err != nil {
				return err
			}
			var node CreateAuthenticatorBeginNode
			if !graph.FindLastNode(&node) {
				// expected there is CreateAuthenticatorBeginNode before the steps
				return webapp.ErrSessionStepMismatch
			}
			switch node.GetCreateAuthenticatorStage() {
			case authn.AuthenticationStagePrimary:
				choiceStep = webapp.SessionStepCreateAuthenticator
				inputFn = func() (interface{}, error) {
					return &InputSelectOOB{}, nil
				}
			case authn.AuthenticationStageSecondary:
				choiceStep = webapp.SessionStepCreateAuthenticator
				// if the user has inputted the phone number when setting up mfa
				// use the selected phone as the input
				// so user doesn't have to input phone number again
				selectedPhone := ""
				var node2 CreateAuthenticatorPhoneOTPNode
				if graph.FindLastNode(&node2) {
					if node2.GetCreateAuthenticatorStage() == authn.AuthenticationStageSecondary {
						selectedPhone = node2.GetSelectedPhoneNumberForPhoneOTP()
					}
				}
				if selectedPhone != "" {
					inputFn = func() (interface{}, error) {
						return &InputSetupOOB{
							InputType: "phone",
							Target:    selectedPhone,
						}, nil
					}
				} else {
					inputFn = nil
				}
			default:
				panic(fmt.Sprintf("webapp: unexpected authentication stage: %s", node.GetCreateAuthenticatorStage()))
			}

		case webapp.SessionStepSetupLoginLinkOTP:
			graph, err := ctrl.InteractionGet(ctx)
			if err != nil {
				return err
			}
			var node CreateAuthenticatorBeginNode
			if !graph.FindLastNode(&node) {
				// expected there is CreateAuthenticatorBeginNode before the steps
				return webapp.ErrSessionStepMismatch
			}
			switch node.GetCreateAuthenticatorStage() {
			case authn.AuthenticationStagePrimary:
				choiceStep = webapp.SessionStepCreateAuthenticator
				inputFn = func() (interface{}, error) {
					return &InputSelectLoginLink{}, nil
				}
			case authn.AuthenticationStageSecondary:
				choiceStep = webapp.SessionStepCreateAuthenticator
			default:
				panic(fmt.Sprintf("webapp: unexpected authentication stage: %s", node.GetCreateAuthenticatorStage()))
			}
		case webapp.SessionStepSetupWhatsappOTP:
			graph, err := ctrl.InteractionGet(ctx)
			if err != nil {
				return err
			}
			var node CreateAuthenticatorBeginNode
			if !graph.FindLastNode(&node) {
				// expected there is CreateAuthenticatorBeginNode before the steps
				return webapp.ErrSessionStepMismatch
			}
			switch node.GetCreateAuthenticatorStage() {
			case authn.AuthenticationStagePrimary:
				choiceStep = webapp.SessionStepCreateAuthenticator
				inputFn = func() (interface{}, error) {
					return &InputSelectWhatsappOTP{}, nil
				}
			case authn.AuthenticationStageSecondary:
				choiceStep = webapp.SessionStepCreateAuthenticator
				// if the user has inputted the phone number when setting up mfa
				// use the selected phone as the input
				// so user doesn't have to input phone number again
				selectedPhone := ""
				var node2 CreateAuthenticatorPhoneOTPNode
				if graph.FindLastNode(&node2) {
					if node2.GetCreateAuthenticatorStage() == authn.AuthenticationStageSecondary {
						selectedPhone = node2.GetSelectedPhoneNumberForPhoneOTP()
					}
				}
				if selectedPhone != "" {
					inputFn = func() (interface{}, error) {
						return &InputSetupWhatsappOTP{
							Phone: selectedPhone,
						}, nil
					}
				} else {
					inputFn = nil
				}
			default:
				panic(fmt.Sprintf("webapp: unexpected authentication stage: %s", node.GetCreateAuthenticatorStage()))
			}
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
		case webapp.SessionStepVerifyWhatsappOTPAuthn:
			choiceStep = webapp.SessionStepAuthenticate
			index, err := strconv.Atoi(ctrl.request.Form.Get("x_authenticator_index"))
			if err != nil {
				index = 0
			}
			inputFn = func() (interface{}, error) {
				return &InputTriggerWhatsApp{
					AuthenticatorIndex: index,
				}, nil
			}
		case webapp.SessionStepVerifyLoginLinkOTPAuthn:
			choiceStep = webapp.SessionStepAuthenticate
			index, err := strconv.Atoi(ctrl.request.Form.Get("x_authenticator_index"))
			if err != nil {
				index = 0
			}
			inputFn = func() (interface{}, error) {
				return &InputTriggerLoginLink{
					AuthenticatorIndex: index,
				}, nil
			}
		case webapp.SessionStepVerifyIdentityViaOOBOTP:
			choiceStep = webapp.SessionStepVerifyIdentityBegin
			inputFn = func() (interface{}, error) {
				return &InputSelectVerifyIdentityViaOOBOTP{}, nil
			}
		case webapp.SessionStepVerifyIdentityViaWhatsapp:
			choiceStep = webapp.SessionStepVerifyIdentityBegin
			inputFn = func() (interface{}, error) {
				return &InputSelectVerifyIdentityViaWhatsapp{}, nil
			}
		case webapp.SessionStepCreatePasskey:
			choiceStep = webapp.SessionStepCreateAuthenticator
			inputFn = func() (interface{}, error) {
				attestationResponseStr := ctrl.request.Form.Get("x_attestation_response")
				attestationResponse := []byte(attestationResponseStr)
				stage := ctrl.request.Form.Get("x_stage")

				return &InputPasskeyAttestationResponse{
					Stage:               stage,
					AttestationResponse: attestationResponse,
				}, nil
			}
		case webapp.SessionStepUsePasskey:
			choiceStep = webapp.SessionStepAuthenticate
			inputFn = func() (interface{}, error) {
				assertionResponseStr := ctrl.request.Form.Get("x_assertion_response")
				assertionResponse := []byte(assertionResponseStr)
				stage := ctrl.request.Form.Get("x_stage")

				return &InputPasskeyAssertionResponse{
					Stage:             stage,
					AssertionResponse: assertionResponse,
				}, nil
			}
			break
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
			if err = ctrl.Page.UpdateSession(ctx, session); err != nil {
				return err
			}
			result = &webapp.Result{
				RedirectURI:      session.CurrentStep().URL().String(),
				NavigationAction: webapp.NavigationActionReplace,
			}
		} else {
			result, err = ctrl.InteractionPost(ctx, inputFn)
			if err != nil {
				return err
			}
			result.NavigationAction = webapp.NavigationActionReplace
		}

		result.WriteResponse(ctrl.response, ctrl.request)
		return nil
	})
}
