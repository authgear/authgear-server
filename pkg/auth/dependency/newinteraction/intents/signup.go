package intents

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
)

func init() {
	newinteraction.RegisterIntent(&IntentSignup{})
}

type IntentSignup struct {
}

func (i *IntentSignup) InstantiateRootNode(ctx *newinteraction.Context, graph *newinteraction.Graph) (newinteraction.Node, error) {
	spec := nodes.EdgeDoCreateUser{}
	return spec.Instantiate(ctx, graph, i)
}

func (i *IntentSignup) DeriveEdgesForNode(ctx *newinteraction.Context, graph *newinteraction.Graph, node newinteraction.Node) ([]newinteraction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeDoCreateUser:
		return []newinteraction.Edge{
			&nodes.EdgeCreateIdentityBegin{},
		}, nil

	case *nodes.NodeCreateIdentityEnd:
		return []newinteraction.Edge{
			&nodes.EdgeCreateAuthenticatorBegin{Stage: newinteraction.AuthenticationStagePrimary},
		}, nil

	case *nodes.NodeCreateAuthenticatorEnd:
		switch node.Stage {
		case newinteraction.AuthenticationStagePrimary:
			return []newinteraction.Edge{
				&nodes.EdgeCreateAuthenticatorBegin{Stage: newinteraction.AuthenticationStageSecondary},
			}, nil
		case newinteraction.AuthenticationStageSecondary:
			// TODO(new_interaction): MFA
			return []newinteraction.Edge{&nodes.EdgeDoCreateSession{Reason: auth.SessionCreateReasonSignup}}, nil
		default:
			panic(fmt.Errorf("interaction: unexpected authentication stage: %v", node.Stage))
		}

	default:
		panic("interaction: unexpected node")
	}
}
