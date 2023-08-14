package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeDoPopulateStandardAttributes{})
}

type NodeDoPopulateStandardAttributes struct {
	Identity *identity.Info `json:"identity,omitempty"`
}

var _ MilestoneDoPopulateStandardAttributes = &NodeDoPopulateStandardAttributes{}

func (*NodeDoPopulateStandardAttributes) Milestone()                             {}
func (*NodeDoPopulateStandardAttributes) MilestoneDoPopulateStandardAttributes() {}

var _ workflow.NodeSimple = &NodeDoPopulateStandardAttributes{}

func (n *NodeDoPopulateStandardAttributes) Kind() string {
	return "workflowconfig.NodeDoPopulateStandardAttributes"
}

func (n *NodeDoPopulateStandardAttributes) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.RunEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
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

func (*NodeDoPopulateStandardAttributes) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDoPopulateStandardAttributes) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeDoPopulateStandardAttributes) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
