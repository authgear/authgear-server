package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeUseIdentityLoginID{})
}

type NodeUseIdentityLoginID struct {
	Identification config.WorkflowIdentificationMethod `json:"identification,omitempty"`
}

func (*NodeUseIdentityLoginID) Kind() string {
	return "workflowconfig.NodeUseIdentityLoginID"
}

func (*NodeUseIdentityLoginID) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return []workflow.Input{&InputTakeLoginID{}}, nil
}

func (n *NodeUseIdentityLoginID) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakeLoginID inputTakeLoginID
	if workflow.AsInput(input, &inputTakeLoginID) {
		loginID := inputTakeLoginID.GetLoginID()
		spec := &identity.Spec{
			Type: model.IdentityTypeLoginID,
			LoginID: &identity.LoginIDSpec{
				Value: loginID,
			},
		}

		exactMatch, otherMatches, err := deps.Identities.SearchBySpec(spec)
		if err != nil {
			return nil, err
		}

		// FIXME(workflow): rate limit on account enumeration

		if exactMatch == nil {
			var otherSpec *identity.Spec
			if len(otherMatches) > 0 {
				s := otherMatches[0].ToSpec()
				otherSpec = &s
			}
			return nil, identityFillDetails(api.ErrUserNotFound, spec, otherSpec)
		}

		return workflow.NewNodeSimple(&NodeDoUseIdentity{
			Identity: exactMatch,
		}), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*NodeUseIdentityLoginID) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return nil, nil
}

func (*NodeUseIdentityLoginID) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
