package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodeDoCheckAccountStatus{})
}

type NodeDoCheckAccountStatus struct {
	UserID string `json:"user_id"`
}

var _ authflow.NodeSimple = &NodeDoCheckAccountStatus{}
var _ authflow.EffectGetter = &NodeDoCheckAccountStatus{}

func (n *NodeDoCheckAccountStatus) Kind() string {
	return "NodeDoCheckAccountStatus"
}

func (n *NodeDoCheckAccountStatus) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]authflow.Effect, error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			u, err := deps.Users.GetRaw(n.UserID)
			if err != nil {
				return err
			}

			err = u.AccountStatus().Check()
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}
