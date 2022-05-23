package viewmodels

import (
	"strconv"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	corephone "github.com/authgear/authgear-server/pkg/util/phone"
)

type AuthenticationPhoneOTPTriggerNode interface {
	GetSelectedPhoneNumberForPhoneOTPAuthentication() string
}

type PhoneOTPAlternativeStepsViewModel struct {
	PhoneOTPAlternativeSteps []AlternativeStep
}

func (m *PhoneOTPAlternativeStepsViewModel) AddAlternatives(graph *interaction.Graph, currentStepKind webapp.SessionStepKind) error {
	// authenticator creation
	var node CreateAuthenticatorBeginNode
	if graph.FindLastNode(&node) {
		return m.addCreateAuthenticatorAlternatives(node, graph, currentStepKind)
	}

	// authentication
	var node2 AuthenticationBeginNode
	if graph.FindLastNode(&node2) {
		return m.addAuthenticationAlternatives(node2, graph, currentStepKind)
	}

	return nil
}

func (m *PhoneOTPAlternativeStepsViewModel) addCreateAuthenticatorAlternatives(node CreateAuthenticatorBeginNode, graph *interaction.Graph, currentStepKind webapp.SessionStepKind) error {
	edges, err := node.GetCreateAuthenticatorEdges()
	if err != nil {
		return err
	}
	for _, edge := range edges {
		switch edge := edge.(type) {
		case *nodes.EdgeCreateAuthenticatorOOBSetup:
			oobType := edge.AuthenticatorType()
			if oobType != model.AuthenticatorTypeOOBSMS {
				continue
			}
			if currentStepKind != webapp.SessionStepSetupOOBOTPSMS &&
				currentStepKind != webapp.SessionStepEnterOOBOTPSetupSMS {
				m.PhoneOTPAlternativeSteps = append(m.PhoneOTPAlternativeSteps, AlternativeStep{
					Step: webapp.SessionStepSetupOOBOTPSMS,
				})
			}
		case *nodes.EdgeCreateAuthenticatorWhatsappOTPSetup:
			if currentStepKind != webapp.SessionStepSetupWhatsappOTP &&
				currentStepKind != webapp.SessionStepVerifyWhatsappOTPSetup {
				m.PhoneOTPAlternativeSteps = append(m.PhoneOTPAlternativeSteps, AlternativeStep{
					Step: webapp.SessionStepSetupWhatsappOTP,
				})
			}
		default:
			continue
		}
	}

	return nil
}

func (m *PhoneOTPAlternativeStepsViewModel) addAuthenticationAlternatives(node AuthenticationBeginNode, graph *interaction.Graph, currentStepKind webapp.SessionStepKind) error {
	edges, err := node.GetAuthenticationEdges()
	if err != nil {
		return err
	}

	var node2 AuthenticationPhoneOTPTriggerNode
	if !graph.FindLastNode(&node2) {
		// PhoneOTPAlternativeStepsViewModel is used by sms otp and whats otp authentication only
		// so it is expected that the graph should has node implementing AuthenticationPhoneOTPTriggerNode
		panic("viewmodels: expected graph has node implementing AuthenticationPhoneOTPTriggerNode")
	}

	// For the whatsapp and sms switches, we only show the authenticator
	// with the same phone number
	// This is different from the AlternativeStepsViewModel
	selectedPhone := node2.GetSelectedPhoneNumberForPhoneOTPAuthentication()

	for _, edge := range edges {
		switch edge := edge.(type) {
		case *nodes.EdgeAuthenticationWhatsappTrigger:
			if currentStepKind != webapp.SessionStepVerifyWhatsappOTPAuthn {
				for i := range edge.Authenticators {
					phone := edge.GetPhone(i)
					if selectedPhone != phone {
						continue
					}
					maskedPhone := corephone.Mask(phone)
					m.PhoneOTPAlternativeSteps = append(m.PhoneOTPAlternativeSteps, AlternativeStep{
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
			if oobAuthenticatorType == model.AuthenticatorTypeOOBSMS &&
				currentStepKind != webapp.SessionStepEnterOOBOTPAuthnSMS {
				show = true
			}

			if !show {
				continue
			}

			for i := range edge.Authenticators {
				target := edge.GetOOBOTPTarget(i)
				maskedTarget := corephone.Mask(target)
				sessionStep := webapp.SessionStepEnterOOBOTPAuthnSMS
				if selectedPhone != target {
					continue
				}

				m.PhoneOTPAlternativeSteps = append(m.PhoneOTPAlternativeSteps, AlternativeStep{
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

		default:

		}
	}
	return nil
}
