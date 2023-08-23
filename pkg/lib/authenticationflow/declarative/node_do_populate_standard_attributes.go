package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeDoPopulateStandardAttributes{})
}

type NodeDoPopulateStandardAttributes struct {
	Identity *identity.Info `json:"identity,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoPopulateStandardAttributes{}
var _ authflow.Milestone = &NodeDoPopulateStandardAttributes{}
var _ MilestoneDoPopulateStandardAttributes = &NodeDoPopulateStandardAttributes{}
var _ authflow.EffectGetter = &NodeDoPopulateStandardAttributes{}

func (n *NodeDoPopulateStandardAttributes) Kind() string {
	return "NodeDoPopulateStandardAttributes"
}

func (*NodeDoPopulateStandardAttributes) Milestone()                             {}
func (*NodeDoPopulateStandardAttributes) MilestoneDoPopulateStandardAttributes() {}

func (n *NodeDoPopulateStandardAttributes) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
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
