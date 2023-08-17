package workflowconfig

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeCreateIdentityLoginID{})
}

type NodeCreateIdentityLoginID struct {
	UserID         string                              `json:"user_id,omitempty"`
	Identification config.WorkflowIdentificationMethod `json:"identification,omitempty"`
}

var _ workflow.NodeSimple = &NodeCreateIdentityLoginID{}
var _ workflow.Milestone = &NodeCreateIdentityLoginID{}
var _ MilestoneIdentificationMethod = &NodeCreateIdentityLoginID{}
var _ workflow.InputReactor = &NodeCreateIdentityLoginID{}

func (*NodeCreateIdentityLoginID) Kind() string {
	return "workflowconfig.NodeCreateIdentityLoginID"
}

func (*NodeCreateIdentityLoginID) Milestone() {}
func (n *NodeCreateIdentityLoginID) MilestoneIdentificationMethod() config.WorkflowIdentificationMethod {
	return n.Identification
}

func (*NodeCreateIdentityLoginID) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return []workflow.Input{&InputTakeLoginID{}}, nil
}

func (n *NodeCreateIdentityLoginID) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakeLoginID inputTakeLoginID
	if workflow.AsInput(input, &inputTakeLoginID) {
		loginID := inputTakeLoginID.GetLoginID()
		spec, err := n.makeLoginIDSpec(loginID)
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

func (n *NodeCreateIdentityLoginID) makeLoginIDSpec(loginID string) (*identity.Spec, error) {
	spec := &identity.Spec{
		Type: model.IdentityTypeLoginID,
		LoginID: &identity.LoginIDSpec{
			Value: loginID,
		},
	}
	switch n.Identification {
	case config.WorkflowIdentificationMethodEmail:
		spec.LoginID.Type = model.LoginIDKeyTypeEmail
		spec.LoginID.Key = string(spec.LoginID.Type)
	case config.WorkflowIdentificationMethodPhone:
		spec.LoginID.Type = model.LoginIDKeyTypePhone
		spec.LoginID.Key = string(spec.LoginID.Type)
	case config.WorkflowIdentificationMethodUsername:
		spec.LoginID.Type = model.LoginIDKeyTypeUsername
		spec.LoginID.Key = string(spec.LoginID.Type)
	default:
		return nil, InvalidIdentificationMethod.New("unexpected identification method")
	}
	return spec, nil
}
