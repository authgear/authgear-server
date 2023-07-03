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
	workflow.RegisterNode(&NodeCreateLoginID{})
}

type NodeCreateLoginID struct {
	UserID         string                              `json:"user_id,omitempty"`
	Identification config.WorkflowIdentificationMethod `json:"identification,omitempty"`
}

func (*NodeCreateLoginID) Kind() string {
	return "workflowconfig.NodeCreateLoginID"
}

func (*NodeCreateLoginID) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	return []workflow.Input{&InputTakeLoginID{}}, nil
}

func (n *NodeCreateLoginID) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
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

func (*NodeCreateLoginID) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (n *NodeCreateLoginID) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}
