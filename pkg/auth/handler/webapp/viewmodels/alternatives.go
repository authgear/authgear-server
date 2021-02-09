package viewmodels

import (
	"fmt"
	"strconv"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	corephone "github.com/authgear/authgear-server/pkg/util/phone"
)

type AuthenticationBeginNode interface {
	GetAuthenticationEdges() ([]interaction.Edge, error)
	GetAuthenticationStage() interaction.AuthenticationStage
}

type CreateAuthenticatorBeginNode interface {
	GetCreateAuthenticatorEdges() ([]interaction.Edge, error)
	GetCreateAuthenticatorStage() interaction.AuthenticationStage
}

type OOBOTPTriggerNode interface {
	GetOOBOTPTarget() string
}

type AlternativeStep struct {
	Step  webapp.SessionStepKind
	Input map[string]string
	Data  map[string]string
}

type AlternativeStepsViewModel struct {
	AuthenticationStage   interaction.AuthenticationStage
	AlternativeSteps      []AlternativeStep
	CanRequestDeviceToken bool
}

func (m *AlternativeStepsViewModel) AddAuthenticationAlternatives(graph *interaction.Graph, currentStepKind webapp.SessionStepKind) error {
	var node AuthenticationBeginNode
	if !graph.FindLastNode(&node) {
		panic("authentication_begin: expected graph has node implementing AuthenticationBeginNode")
	}

	m.AuthenticationStage = node.GetAuthenticationStage()

	edges, err := node.GetAuthenticationEdges()
	if err != nil {
		return err
	}

	for _, edge := range edges {
		switch edge := edge.(type) {
		case *nodes.EdgeUseDeviceToken:
			m.CanRequestDeviceToken = true
		case *nodes.EdgeConsumeRecoveryCode:
			if currentStepKind != webapp.SessionStepEnterRecoveryCode {
				m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
					Step: webapp.SessionStepEnterRecoveryCode,
				})
			}
		case *nodes.EdgeAuthenticationPassword:
			if currentStepKind != webapp.SessionStepEnterPassword {
				m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
					Step: webapp.SessionStepEnterPassword,
				})
			}
		case *nodes.EdgeAuthenticationTOTP:
			if currentStepKind != webapp.SessionStepEnterTOTP {
				m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
					Step: webapp.SessionStepEnterTOTP,
				})
			}
		case *nodes.EdgeAuthenticationOOBTrigger:
			show := false
			oobAuthenticatorType := edge.OOBAuthenticatorType
			if oobAuthenticatorType == authn.AuthenticatorTypeOOBSMS &&
				currentStepKind != webapp.SessionStepEnterOOBOTPAuthnSMS {
				show = true
			}

			if oobAuthenticatorType == authn.AuthenticatorTypeOOBEmail &&
				currentStepKind != webapp.SessionStepEnterOOBOTPAuthnEmail {
				show = true
			}

			if show {
				currentTarget := ""
				var node OOBOTPTriggerNode
				if graph.FindLastNode(&node) {
					currentTarget = node.GetOOBOTPTarget()
				}

				for i := range edge.Authenticators {
					target := edge.GetOOBOTPTarget(i)

					var maskedTarget string
					var sessionStep webapp.SessionStepKind
					switch oobAuthenticatorType {
					case authn.AuthenticatorTypeOOBSMS:
						maskedTarget = corephone.Mask(target)
						sessionStep = webapp.SessionStepEnterOOBOTPAuthnSMS
					case authn.AuthenticatorTypeOOBEmail:
						maskedTarget = mail.MaskAddress(target)
						sessionStep = webapp.SessionStepEnterOOBOTPAuthnEmail
					default:
						panic("authentication_begin: unexpected oob authenticator type: " + oobAuthenticatorType)
					}

					if currentTarget == target {
						continue
					}

					m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
						Step: sessionStep,
						Input: map[string]string{
							"x_authenticator_type":  string(oobAuthenticatorType),
							"x_authenticator_index": strconv.Itoa(i),
						},
						Data: map[string]string{
							"target": maskedTarget,
						},
					})
				}
			}
		default:
			panic(fmt.Errorf("authentication_begin: unexpected edge: %T", edge))
		}
	}
	return nil
}

func (m *AlternativeStepsViewModel) AddCreateAuthenticatorAlternatives(graph *interaction.Graph, currentStepKind webapp.SessionStepKind) error {
	var node CreateAuthenticatorBeginNode
	if !graph.FindLastNode(&node) {
		panic("create_authenticator_begin: expected graph has node implementing CreateAuthenticatorBeginNode")
	}

	m.AuthenticationStage = node.GetCreateAuthenticatorStage()

	edges, err := node.GetCreateAuthenticatorEdges()
	if err != nil {
		return err
	}

	for _, edge := range edges {
		switch edge := edge.(type) {
		case *nodes.EdgeCreateAuthenticatorPassword:
			if currentStepKind != webapp.SessionStepCreatePassword {
				m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
					Step: webapp.SessionStepCreatePassword,
				})
			}
		case *nodes.EdgeCreateAuthenticatorOOBSetup:
			oobType := edge.AuthenticatorType()
			switch oobType {
			case authn.AuthenticatorTypeOOBEmail:
				if currentStepKind != webapp.SessionStepSetupOOBOTPEmail &&
					currentStepKind != webapp.SessionStepEnterOOBOTPSetupEmail {
					m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
						Step: webapp.SessionStepSetupOOBOTPEmail,
					})
				}
			case authn.AuthenticatorTypeOOBSMS:
				if currentStepKind != webapp.SessionStepSetupOOBOTPSMS &&
					currentStepKind != webapp.SessionStepEnterOOBOTPSetupSMS {
					m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
						Step: webapp.SessionStepSetupOOBOTPSMS,
					})
				}
			default:
				panic(fmt.Errorf("create_authenticator_begin: authenticator type in oob edge: %s", oobType))
			}
		case *nodes.EdgeCreateAuthenticatorTOTPSetup:
			if currentStepKind != webapp.SessionStepSetupTOTP {
				m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
					Step: webapp.SessionStepSetupTOTP,
				})
			}
		default:
			panic(fmt.Errorf("create_authenticator_begin: unexpected edge: %T", edge))
		}
	}

	return nil
}
