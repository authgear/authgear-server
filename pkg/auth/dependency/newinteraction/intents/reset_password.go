package intents

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
)

func init() {
	newinteraction.RegisterIntent(&IntentResetPassword{})
}

type IntentResetPassword struct{}

func NewIntentResetPassword() *IntentResetPassword {
	return &IntentResetPassword{}
}

func (i *IntentResetPassword) InstantiateRootNode(ctx *newinteraction.Context, graph *newinteraction.Graph) (newinteraction.Node, error) {
	return &nodes.NodeResetPasswordBegin{}, nil
}

func (i *IntentResetPassword) DeriveEdgesForNode(ctx *newinteraction.Context, graph *newinteraction.Graph, node newinteraction.Node) ([]newinteraction.Edge, error) {
	return nil, nil
}
