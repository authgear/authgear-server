package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeDoPopulateStandardAttributesInSignup{})
}

// This node is only for use in signup or promote
type NodeDoPopulateStandardAttributesInSignup struct {
	Identity *identity.Info `json:"identity,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoPopulateStandardAttributesInSignup{}
var _ authflow.Milestone = &NodeDoPopulateStandardAttributesInSignup{}
var _ MilestoneDoPopulateStandardAttributes = &NodeDoPopulateStandardAttributesInSignup{}
var _ authflow.EffectGetter = &NodeDoPopulateStandardAttributesInSignup{}
var _ MilestoneSwitchToExistingUser = &NodeDoPopulateStandardAttributesInSignup{}

func (n *NodeDoPopulateStandardAttributesInSignup) Kind() string {
	return "NodeDoPopulateStandardAttributes"
}

func (*NodeDoPopulateStandardAttributesInSignup) Milestone() {}

func (*NodeDoPopulateStandardAttributesInSignup) MilestoneDoPopulateStandardAttributes() {}
func (i *NodeDoPopulateStandardAttributesInSignup) MilestoneSwitchToExistingUser(newUserID string) {
	// TODO(tung): Skip this step
	i.Identity = i.Identity.UpdateUserID(newUserID)
}

func (n *NodeDoPopulateStandardAttributesInSignup) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			if deps.Config.UserProfile.StandardAttributes.Population.Strategy == config.StandardAttributesPopulationStrategyOnSignup {
				err := deps.StdAttrsService.PopulateStandardAttributes(
					n.Identity.UserID,
					n.Identity,
				)
				if err != nil {
					return err
				}
			}

			return nil
		}),
	}, nil
}
