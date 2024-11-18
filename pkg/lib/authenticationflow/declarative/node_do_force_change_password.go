package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

func init() {
	authflow.RegisterNode(&NodeDoForceChangePassword{})
}

type NodeDoForceChangePassword struct {
	Authenticator *authenticator.Info   `json:"authenticator,omitempty"`
	Reason        *PasswordChangeReason `json:"reason,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoForceChangePassword{}
var _ authflow.EffectGetter = &NodeDoForceChangePassword{}

func (*NodeDoForceChangePassword) Kind() string {
	return "NodeDoForceChangePassword"
}

func (n *NodeDoForceChangePassword) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]authflow.Effect, error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.Authenticators.Update(ctx, n.Authenticator)
		}),
		// authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
		// 	reason := ""
		// 	if n.Reason != nil {
		// 		reason = string(*n.Reason)
		// 	}
		//
		// 	switch n.Authenticator.Kind {
		// 	case authenticator.KindPrimary:
		// 		err := deps.Events.DispatchEventOnCommit(ctx, &nonblocking.PasswordPrimaryForceChangedEventPayload{
		// 			UserRef: model.UserRef{
		// 				Meta: model.Meta{
		// 					ID: n.Authenticator.UserID,
		// 				},
		// 			},
		// 			Reason: reason,
		// 		})
		// 		if err != nil {
		// 			return err
		// 		}
		// 		return nil
		// 	case authenticator.KindSecondary:
		// 		err := deps.Events.DispatchEventOnCommit(ctx, &nonblocking.PasswordSecondaryForceChangedEventPayload{
		// 			UserRef: model.UserRef{
		// 				Meta: model.Meta{
		// 					ID: n.Authenticator.UserID,
		// 				},
		// 			},
		// 			Reason: reason,
		// 		})
		// 		if err != nil {
		// 			return err
		// 		}
		// 		return nil
		// 	default:
		// 		panic(fmt.Errorf("unexpected authenticator kind: %v", n.Authenticator.Kind))
		// 	}
		// }),
	}, nil
}
