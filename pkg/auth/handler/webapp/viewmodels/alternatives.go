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
	AlternativeSteps []AlternativeStep
}

func (m *AlternativeStepsViewModel) AddAuthenticationAlternatives(graph *interaction.Graph, currentStepKind webapp.SessionStepKind) error {
	var node AuthenticationBeginNode
	if !graph.FindLastNode(&node) {
		panic("authentication_begin: expected graph has node implementing AuthenticationBeginNode")
	}

	edges, err := node.GetAuthenticationEdges()
	if err != nil {
		return err
	}

	for _, edge := range edges {
		switch edge := edge.(type) {
		case *nodes.EdgeUseDeviceToken:
			break
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
			if currentStepKind != webapp.SessionStepEnterOOBOTPAuthn {
				currentTarget := ""
				var node OOBOTPTriggerNode
				if graph.FindLastNode(&node) {
					currentTarget = node.GetOOBOTPTarget()
				}

				for i := range edge.Authenticators {
					channel := edge.GetOOBOTPChannel(i)
					target := edge.GetOOBOTPTarget(i)

					var maskedTarget string
					switch channel {
					case string(authn.AuthenticatorOOBChannelSMS):
						maskedTarget = corephone.Mask(target)
					case string(authn.AuthenticatorOOBChannelEmail):
						maskedTarget = mail.MaskAddress(target)
					default:
						panic("authentication_begin: unexpected channel: " + channel)
					}

					if currentTarget == target {
						continue
					}

					m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
						Step: webapp.SessionStepEnterOOBOTPAuthn,
						Input: map[string]string{
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

	edges, err := node.GetCreateAuthenticatorEdges()
	if err != nil {
		return err
	}

	for _, edge := range edges {
		switch edge.(type) {
		case *nodes.EdgeCreateAuthenticatorPassword:
			if currentStepKind != webapp.SessionStepCreatePassword {
				m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
					Step: webapp.SessionStepCreatePassword,
				})
			}
		case *nodes.EdgeCreateAuthenticatorOOBSetup:
			if currentStepKind != webapp.SessionStepSetupOOBOTP {
				m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
					Step: webapp.SessionStepSetupOOBOTP,
				})
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
