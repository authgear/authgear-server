package viewmodels

import (
	"fmt"
	"strconv"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	corephone "github.com/authgear/authgear-server/pkg/util/phone"
)

type AuthenticationBeginNode interface {
	GetAuthenticationEdges() ([]interaction.Edge, error)
	GetAuthenticationStage() authn.AuthenticationStage
}

type CreateAuthenticatorBeginNode interface {
	GetCreateAuthenticatorEdges() ([]interaction.Edge, error)
	GetCreateAuthenticatorStage() authn.AuthenticationStage
}

type OOBOTPTriggerNode interface {
	GetOOBOTPTarget() string
}

type LoginLinkTriggerNode interface {
	GetLoginLinkOTPTarget() string
}

type WhatsappOTPTriggerNode interface {
	GetPhone() string
}

type AlternativeStep struct {
	Step  webapp.SessionStepKind
	Input map[string]string
	Data  map[string]string
}

type AlternativeStepsViewModel struct {
	AuthenticationStage   string
	AlternativeSteps      []AlternativeStep
	CanRequestDeviceToken bool
}

type AlternativeStepsViewModeler struct {
	AuthenticationConfig *config.AuthenticationConfig
}

// nolint: gocognit
func (a *AlternativeStepsViewModeler) AuthenticationAlternatives(graph *interaction.Graph, currentStepKind webapp.SessionStepKind) (*AlternativeStepsViewModel, error) {
	m := &AlternativeStepsViewModel{}

	var node AuthenticationBeginNode
	if !graph.FindLastNode(&node) {
		panic("authentication_begin: expected graph has node implementing AuthenticationBeginNode")
	}

	m.AuthenticationStage = string(node.GetAuthenticationStage())

	edges, err := node.GetAuthenticationEdges()
	if err != nil {
		return nil, err
	}

	phoneOTPStepAdded := false
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
		case *nodes.EdgeAuthenticationPasskey:
			if currentStepKind != webapp.SessionStepUsePasskey {
				m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
					Step: webapp.SessionStepUsePasskey,
				})
			}
		case *nodes.EdgeAuthenticationTOTP:
			if currentStepKind != webapp.SessionStepEnterTOTP {
				m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
					Step: webapp.SessionStepEnterTOTP,
				})
			}
		case *nodes.EdgeAuthenticationLoginLinkTrigger:
			if currentStepKind != webapp.SessionStepEnterOOBOTPAuthnEmail {
				currentTarget := ""
				var node LoginLinkTriggerNode
				if graph.FindLastNode(&node) {
					currentTarget = node.GetLoginLinkOTPTarget()
				}

				for i := range edge.Authenticators {
					target := edge.GetTarget(i)
					if currentTarget == target {
						continue
					}
					m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
						Step: webapp.SessionStepVerifyLoginLinkOTPAuthn,
						Input: map[string]string{
							"x_authenticator_index": strconv.Itoa(i),
						},
						Data: map[string]string{
							"target": mail.MaskAddress(target),
						},
					})
				}
			}
		case *nodes.EdgeAuthenticationWhatsappTrigger:
			if !phoneOTPStepAdded &&
				currentStepKind != webapp.SessionStepEnterOOBOTPAuthnSMS &&
				currentStepKind != webapp.SessionStepVerifyWhatsappOTPAuthn {
				phoneOTPStepAdded = true

				currentPhone := ""
				var node WhatsappOTPTriggerNode
				if graph.FindLastNode(&node) {
					currentPhone = node.GetPhone()
				}

				for i := range edge.Authenticators {
					phone := edge.GetPhone(i)
					if currentPhone == phone {
						continue
					}
					maskedPhone := corephone.Mask(phone)
					m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
						Step: webapp.SessionStepVerifyWhatsappOTPAuthn,
						Input: map[string]string{
							"x_authenticator_index": strconv.Itoa(i),
						},
						Data: map[string]string{
							"target": maskedPhone,
						},
					})
				}
			}
		case *nodes.EdgeAuthenticationOOBTrigger:
			show := false
			oobAuthenticatorType := edge.OOBAuthenticatorType
			if !phoneOTPStepAdded &&
				oobAuthenticatorType == model.AuthenticatorTypeOOBSMS &&
				currentStepKind != webapp.SessionStepEnterOOBOTPAuthnSMS &&
				currentStepKind != webapp.SessionStepVerifyWhatsappOTPAuthn {
				show = true
				phoneOTPStepAdded = true
			}

			if oobAuthenticatorType == model.AuthenticatorTypeOOBEmail &&
				currentStepKind != webapp.SessionStepEnterOOBOTPAuthnEmail &&
				currentStepKind != webapp.SessionStepVerifyLoginLinkOTPAuthn {
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
					case model.AuthenticatorTypeOOBSMS:
						maskedTarget = corephone.Mask(target)
						sessionStep = webapp.SessionStepEnterOOBOTPAuthnSMS
					case model.AuthenticatorTypeOOBEmail:
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

	return m, nil
}

// nolint: gocognit
func (a *AlternativeStepsViewModeler) CreateAuthenticatorAlternatives(graph *interaction.Graph, currentStepKind webapp.SessionStepKind) (*AlternativeStepsViewModel, error) {
	m := &AlternativeStepsViewModel{}

	var node CreateAuthenticatorBeginNode
	if !graph.FindLastNode(&node) {
		panic("create_authenticator_begin: expected graph has node implementing CreateAuthenticatorBeginNode")
	}

	m.AuthenticationStage = string(node.GetCreateAuthenticatorStage())

	edges, err := node.GetCreateAuthenticatorEdges()
	if err != nil {
		return nil, err
	}

	phoneOTPStepAdded := false
	for _, edge := range edges {
		switch edge := edge.(type) {
		case *nodes.EdgeCreateAuthenticatorPassword:
			if currentStepKind != webapp.SessionStepCreatePassword {
				m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
					Step: webapp.SessionStepCreatePassword,
				})
			}
		case *nodes.EdgeCreateAuthenticatorPasskey:
			if currentStepKind != webapp.SessionStepCreatePasskey {
				m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
					Step: webapp.SessionStepCreatePasskey,
				})
			}
		case *nodes.EdgeCreateAuthenticatorOOBSetup:
			oobType := edge.AuthenticatorType()
			switch oobType {
			case model.AuthenticatorTypeOOBEmail:
				if currentStepKind != webapp.SessionStepSetupOOBOTPEmail &&
					currentStepKind != webapp.SessionStepEnterOOBOTPSetupEmail &&
					currentStepKind != webapp.SessionStepSetupLoginLinkOTP &&
					currentStepKind != webapp.SessionStepVerifyLoginLinkOTPSetup {
					m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
						Step: webapp.SessionStepSetupOOBOTPEmail,
					})
				}
			case model.AuthenticatorTypeOOBSMS:
				if !phoneOTPStepAdded &&
					currentStepKind != webapp.SessionStepSetupOOBOTPSMS &&
					currentStepKind != webapp.SessionStepEnterOOBOTPSetupSMS &&
					currentStepKind != webapp.SessionStepSetupWhatsappOTP &&
					currentStepKind != webapp.SessionStepVerifyWhatsappOTPSetup {
					phoneOTPStepAdded = true
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
		case *nodes.EdgeCreateAuthenticatorWhatsappOTPSetup:
			if !phoneOTPStepAdded &&
				currentStepKind != webapp.SessionStepSetupWhatsappOTP &&
				currentStepKind != webapp.SessionStepVerifyWhatsappOTPSetup &&
				currentStepKind != webapp.SessionStepSetupOOBOTPSMS &&
				currentStepKind != webapp.SessionStepEnterOOBOTPSetupSMS {
				phoneOTPStepAdded = true
				m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
					Step: webapp.SessionStepSetupWhatsappOTP,
				})
			}
		case *nodes.EdgeCreateAuthenticatorLoginLinkOTPSetup:
			if currentStepKind != webapp.SessionStepSetupLoginLinkOTP &&
				currentStepKind != webapp.SessionStepSetupOOBOTPEmail &&
				currentStepKind != webapp.SessionStepEnterOOBOTPSetupEmail {
				m.AlternativeSteps = append(m.AlternativeSteps, AlternativeStep{
					Step: webapp.SessionStepSetupLoginLinkOTP,
				})
			}
		default:
			panic(fmt.Errorf("create_authenticator_begin: unexpected edge: %T", edge))
		}
	}

	return m, nil
}
