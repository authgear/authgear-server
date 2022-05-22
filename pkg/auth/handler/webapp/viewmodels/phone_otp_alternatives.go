package viewmodels

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

type PhoneOTPAlternativeStepsViewModel struct {
	PhoneOTPAlternativeSteps []AlternativeStep
}

func (m *PhoneOTPAlternativeStepsViewModel) AddAlternatives(graph *interaction.Graph, currentStepKind webapp.SessionStepKind) error {
	// authenticator creation
	var node CreateAuthenticatorBeginNode
	if graph.FindLastNode(&node) {
		return m.addCreateAuthenticatorAlternatives(node, graph, currentStepKind)
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
