package intents

import (
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

func init() {
	interaction.RegisterIntent(&IntentForgotPassword{})
}

type IntentForgotPassword struct{}

func NewIntentForgotPassword() *IntentForgotPassword {
	return &IntentForgotPassword{}
}

func (i *IntentForgotPassword) InstantiateRootNode(ctx *interaction.Context, graph *interaction.Graph) (interaction.Node, error) {
	return &nodes.NodeForgotPasswordBegin{}, nil
}

func (i *IntentForgotPassword) DeriveEdgesForNode(graph *interaction.Graph, node interaction.Node) ([]interaction.Edge, error) {
	return nil, nil
}
