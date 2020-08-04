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
	Stage newinteraction.AuthenticationStage
}

func (e *EdgeCreateAuthenticatorBegin) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeCreateAuthenticatorBegin{
		Stage: e.Stage,
	}, nil
}

type NodeCreateAuthenticatorBegin struct {
	Stage newinteraction.AuthenticationStage `json:"stage"`
}

func (n *NodeCreateAuthenticatorBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorBegin) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	var edges []newinteraction.Edge

	var requiredType authn.AuthenticatorType
	switch n.Stage {
	case newinteraction.AuthenticationStagePrimary:
		iden := graph.MustGetUserLastIdentity()
		primaryAuthenticatorTypes := iden.Type.PrimaryAuthenticatorTypes()

		ais, err := ctx.Authenticators.List(
			iden.UserID,
			authenticator.KeepTag(authenticator.TagPrimaryAuthenticator),
			authenticator.KeepPrimaryAuthenticatorOfIdentity(iden),
		)
		if err != nil {
			return nil, err
		}

		if len(primaryAuthenticatorTypes) > 0 && len(ctx.Config.Authentication.PrimaryAuthenticators) > 0 {
			first := ctx.Config.Authentication.PrimaryAuthenticators[0]

			found := false
			for _, ai := range ais {
				if ai.Type == first {
					found = true
					break
				}
			}

			if !found {
				requiredType = first
			}
		}
	case newinteraction.AuthenticationStageSecondary:
		// TODO(new_interaction): MFA
		break
	default:
		panic("interaction: unknown authentication stage: " + n.Stage)
	}

	if requiredType != "" {
		switch requiredType {
		case authn.AuthenticatorTypePassword:
			edges = append(edges, &EdgeCreateAuthenticatorPassword{Stage: n.Stage})

		case authn.AuthenticatorTypeTOTP:
			edges = append(edges, &EdgeCreateAuthenticatorTOTPSetup{Stage: n.Stage})

		case authn.AuthenticatorTypeOOB:
			edges = append(edges, &EdgeCreateAuthenticatorOOBSetup{Stage: n.Stage})
		default:
			// TODO(new_interaction): implements bearer token, recovery code
			panic("interaction: unknown authenticator type: " + requiredType)
		}
	}

	// No authenticators needed, skip the stage
	if len(edges) == 0 {
		edges = append(edges, &EdgeCreateAuthenticatorEnd{Stage: n.Stage, Authenticators: nil})
	}

	// TODO(interaction): support choosing authenticator to create
	return edges[:1], nil
}
