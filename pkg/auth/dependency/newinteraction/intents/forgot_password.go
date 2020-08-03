package intents

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
)

func init() {
	newinteraction.RegisterIntent(&IntentForgotPassword{})
}

type IntentForgotPassword struct{}

func NewIntentForgotPassword() *IntentForgotPassword {
	return &IntentForgotPassword{}
}

func (i *IntentForgotPassword) InstantiateRootNode(ctx *newinteraction.Context, graph *newinteraction.Graph) (newinteraction.Node, error) {
	return &nodes.NodeForgotPasswordBegin{}, nil
}

func (i *IntentForgotPassword) DeriveEdgesForNode(ctx *newinteraction.Context, graph *newinteraction.Graph, node newinteraction.Node) ([]newinteraction.Edge, error) {
	return nil, nil
}
