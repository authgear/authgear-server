package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeDoUpdateUserProfile{})
}

type NodeDoUpdateUserProfile struct {
	UserID             string     `json:"user_id,omitempty"`
	StandardAttributes attrs.List `json:"standard_attributes,omitempty"`
	CustomAttributes   attrs.List `json:"custom_attributes,omitempty"`
}

func (*NodeDoUpdateUserProfile) Kind() string {
	return "workflowconfig.NodeDoUpdateUserProfile"
}

func (n *NodeDoUpdateUserProfile) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.RunEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			// FIXME(workflow): support other role?
			err := deps.StdAttrsService.UpdateStandardAttributesWithList(config.RoleEndUser, n.UserID, n.StandardAttributes)
			if err != nil {
				return err
			}
			// FIXME(workflow): support other role?
			err = deps.CustomAttrsService.UpdateCustomAttributesWithList(config.RoleEndUser, n.UserID, n.CustomAttributes)
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}

func (*NodeDoUpdateUserProfile) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDoUpdateUserProfile) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (*NodeDoUpdateUserProfile) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
