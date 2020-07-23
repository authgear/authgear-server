package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

type EdgeAuthenticationBegin struct {
	Stage newinteraction.AuthenticationStage
}

func (e *EdgeAuthenticationBegin) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeAuthenticationBegin{
		Stage: e.Stage,
	}, nil
}

type NodeAuthenticationBegin struct {
	Stage newinteraction.AuthenticationStage `json:"stage"`
}

func (n *NodeAuthenticationBegin) Apply(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeAuthenticationBegin) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
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
		switch t {
		case authn.AuthenticatorTypePassword:
			edges = append(edges, &EdgeAuthenticationPassword{})

		case authn.AuthenticatorTypeTOTP:
			_, infos, err := getAuthenticators(ctx, graph, n.Stage, authn.AuthenticatorTypeTOTP)
			if err != nil {
				return nil, err
			}

			if len(infos) > 0 {
				edges = append(edges, &EdgeAuthenticationTOTP{})
			}

		case authn.AuthenticatorTypeOOB:
			_, infos, err := getAuthenticators(ctx, graph, n.Stage, authn.AuthenticatorTypeOOB)
			if err != nil {
				return nil, err
			}

			if len(infos) > 0 {
				edges = append(edges, &EdgeAuthenticationOOBTrigger{})
			}

		default:
			// TODO(new_interaction): implements bearer token, recovery code
			panic("interaction: unknown identity type: " + t)
		}
	}

	return edges, nil
}
