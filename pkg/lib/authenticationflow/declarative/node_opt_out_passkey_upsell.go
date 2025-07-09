package declarative

import (
	context "context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

type NodeOptOutPasskeyUpsell struct {
	UserID string `json:"user_id"`
}

var _ authflow.NodeSimple = &NodeOptOutPasskeyUpsell{}
var _ authflow.EffectGetter = &NodeOptOutPasskeyUpsell{}

func (*NodeOptOutPasskeyUpsell) Kind() string {
	return "NodeOptOutPasskeyUpsell"
}

func (n *NodeOptOutPasskeyUpsell) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]authflow.Effect, error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.Users.UpdateOptOutPasskeyUpselling(ctx, n.UserID, true)
		}),
	}, nil
}
