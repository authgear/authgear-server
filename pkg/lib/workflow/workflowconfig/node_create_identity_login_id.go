package workflowconfig

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeCreateIdentityLoginID{})
}

type NodeCreateIdentityLoginID struct {
	UserID         string                              `json:"user_id,omitempty"`
	Identification config.WorkflowIdentificationMethod `json:"identification,omitempty"`
}

func (*NodeCreateIdentityLoginID) Kind() string {
	return "workflowconfig.NodeCreateIdentityLoginID"
}

func (*NodeCreateIdentityLoginID) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return []workflow.Input{&InputTakeLoginID{}}, nil
}

func (n *NodeCreateIdentityLoginID) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakeLoginID inputTakeLoginID
	if workflow.AsInput(input, &inputTakeLoginID) {
		loginID := inputTakeLoginID.GetLoginID()
		spec, err := MakeLoginIDSpec(n.Identification, loginID)
		if err != nil {
			return nil, err
		}

		// FIXME(workflow): allow bypassing email blocklist for Admin API.
		info, err := deps.Identities.New(n.UserID, spec, identity.NewIdentityOptions{})
		if err != nil {
			return nil, err
		}

		duplicate, err := deps.Identities.CheckDuplicated(info)
		if err != nil && !errors.Is(err, identity.ErrIdentityAlreadyExists) {
			return nil, err
		}

		if err != nil {
			spec := info.ToSpec()
			otherSpec := duplicate.ToSpec()
			return nil, identityFillDetails(api.ErrDuplicatedIdentity, &spec, &otherSpec)
		}

		return workflow.NewNodeSimple(&NodeDoCreateIdentity{
			Identity: info,
		}), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*NodeCreateIdentityLoginID) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (n *NodeCreateIdentityLoginID) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
