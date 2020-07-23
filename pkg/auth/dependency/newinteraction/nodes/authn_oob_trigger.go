package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

type InputAuthenticationOOBTrigger interface {
	GetOOBTarget() string
}

type EdgeAuthenticationOOBTrigger struct {
	Stage newinteraction.AuthenticationStage
}

func (e *EdgeAuthenticationOOBTrigger) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputAuthenticationOOBTrigger)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	infos, err := getAuthenticators(ctx, graph, e.Stage, authn.AuthenticatorTypeOOB)
	if err != nil {
		return nil, err
	}

	var targetInfo *authenticator.Info
	if len(infos) == 1 {
		// Select the only authenticator by default
		targetInfo = infos[0]
	} else if target := input.GetOOBTarget(); target != "" {
		// Match authenticators with OOB target
		for _, info := range infos {
			switch info.Props[authenticator.AuthenticatorPropOOBOTPChannelType] {
			case authn.AuthenticatorOOBChannelEmail:
				if info.Props[authenticator.AuthenticatorPropOOBOTPEmail] == target {
					targetInfo = info
				}
			case authn.AuthenticatorOOBChannelSMS:
				if info.Props[authenticator.AuthenticatorPropOOBOTPPhone] == target {
					targetInfo = info
				}
			}
		}
	}

	if targetInfo == nil {
		return nil, ErrOOBTargetNotFound
	}

	// TODO(new_interaction): generate code & trigger

	return &NodeAuthenticationOOBTrigger{Stage: e.Stage, Authenticator: targetInfo}, nil
}

type NodeAuthenticationOOBTrigger struct {
	Stage         newinteraction.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info                `json:"authenticator"`
}

func (n *NodeAuthenticationOOBTrigger) Apply(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeAuthenticationOOBTrigger) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{
		&EdgeAuthenticationEnd{Stage: n.Stage, Authenticator: n.Authenticator},
	}, nil
}
