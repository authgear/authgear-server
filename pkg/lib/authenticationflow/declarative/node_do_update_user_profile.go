package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeDoUpdateUserProfile{})
}

type NodeDoUpdateUserProfile struct {
	UserID             string     `json:"user_id,omitempty"`
	StandardAttributes attrs.List `json:"standard_attributes,omitempty"`
	CustomAttributes   attrs.List `json:"custom_attributes,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoUpdateUserProfile{}
var _ authflow.EffectGetter = &NodeDoUpdateUserProfile{}
var _ authflow.Milestone = &NodeDoUpdateUserProfile{}
var _ MilestoneSwitchToExistingUser = &NodeDoUpdateUserProfile{}

func (*NodeDoUpdateUserProfile) Kind() string {
	return "NodeDoUpdateUserProfile"
}

func (*NodeDoUpdateUserProfile) Milestone() {}
func (i *NodeDoUpdateUserProfile) MilestoneSwitchToExistingUser(newUserID string) {
	// TODO(tung): Skip this step
	i.UserID = newUserID
}

func (n *NodeDoUpdateUserProfile) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			// FIXME(authflow): support other role?
			err := deps.StdAttrsService.UpdateStandardAttributesWithList(config.RoleEndUser, n.UserID, n.StandardAttributes)
			if err != nil {
				return err
			}
			// FIXME(authflow): support other role?
			err = deps.CustomAttrsService.UpdateCustomAttributesWithList(config.RoleEndUser, n.UserID, n.CustomAttributes)
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}
