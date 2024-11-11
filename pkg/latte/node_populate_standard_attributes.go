package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodePopulateStandardAttributes{})
}

type NodePopulateStandardAttributes struct {
	Identity *identity.Info `json:"identity,omitempty"`
}

func (n *NodePopulateStandardAttributes) Kind() string {
	return "latte.NodePopulateStandardAttributes"
}

func (n *NodePopulateStandardAttributes) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.RunEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			if deps.Config.UserProfile.StandardAttributes.Population.Strategy == config.StandardAttributesPopulationStrategyOnSignup {
				err := deps.StdAttrsService.PopulateStandardAttributes(ctx,
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

func (*NodePopulateStandardAttributes) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodePopulateStandardAttributes) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodePopulateStandardAttributes) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
