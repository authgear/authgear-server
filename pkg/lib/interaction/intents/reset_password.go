package intents

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

func init() {
	interaction.RegisterIntent(&IntentResetPassword{})
}

type IntentResetPassword struct{}

func NewIntentResetPassword() *IntentResetPassword {
	return &IntentResetPassword{}
}

func (i *IntentResetPassword) InstantiateRootNode(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) (interaction.Node, error) {
	return &nodes.NodeResetPasswordBegin{}, nil
}

func (i *IntentResetPassword) DeriveEdgesForNode(goCtx context.Context, graph *interaction.Graph, node interaction.Node) ([]interaction.Edge, error) {
	return nil, nil
}
