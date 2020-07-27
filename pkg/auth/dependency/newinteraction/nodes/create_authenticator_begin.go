package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeCreateAuthenticatorBegin{})
}

type EdgeCreateAuthenticatorBegin struct {
	Stage                  newinteraction.AuthenticationStage
	RequestedAuthenticator *authenticator.Spec
}

func (e *EdgeCreateAuthenticatorBegin) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeCreateAuthenticatorBegin{
		Stage: e.Stage,
	}, nil
}

type NodeCreateAuthenticatorBegin struct {
	Stage                  newinteraction.AuthenticationStage `json:"stage"`
	RequestedAuthenticator *authenticator.Spec                `json:"requested_authenticator"`
}

func (n *NodeCreateAuthenticatorBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorBegin) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	var edges []newinteraction.Edge

	var availableAuthenticators []authn.AuthenticatorType
	switch n.Stage {
	case newinteraction.AuthenticationStagePrimary:
		availableAuthenticators = ctx.Config.Authentication.PrimaryAuthenticators
	case newinteraction.AuthenticationStageSecondary:
		availableAuthenticators = ctx.Config.Authentication.SecondaryAuthenticators
	default:
		panic("interaction: unknown authentication stage: " + n.Stage)
	}

	for _, t := range availableAuthenticators {
		if n.RequestedAuthenticator != nil && n.RequestedAuthenticator.Type != t {
			continue
		}

		switch t {
		case authn.AuthenticatorTypePassword:
			edges = append(edges, &EdgeCreateAuthenticatorPassword{Stage: n.Stage})

		case authn.AuthenticatorTypeTOTP:
			edges = append(edges, &EdgeCreateAuthenticatorTOTPSetup{Stage: n.Stage})

		case authn.AuthenticatorTypeOOB:
			edges = append(edges, &EdgeCreateAuthenticatorOOBSetup{Stage: n.Stage})

		default:
			// TODO(new_interaction): implements bearer token, recovery code
			panic("interaction: unknown authenticator type: " + t)
		}
	}

	// No authenticators needed, skip the stage
	if len(edges) == 0 {
		edges = append(edges, &EdgeCreateAuthenticatorEnd{Stage: n.Stage, Authenticators: nil})
	}

	return edges, nil
}
