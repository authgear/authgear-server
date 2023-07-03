package workflowconfig

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
	return "workflowconfig.NodePopulateStandardAttributes"
}

func (n *NodePopulateStandardAttributes) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
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

func (*NodePopulateStandardAttributes) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodePopulateStandardAttributes) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodePopulateStandardAttributes) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}
