package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/stringutil"
)

func init() {
	workflow.RegisterNode(&NodeChangeEmail{})
}

type NodeChangeEmail struct {
	UserID               string         `json:"user_id,omitempty"`
	IdentityBeforeUpdate *identity.Info `json:"identity_before_update,omitempty"`
}

func (n *NodeChangeEmail) Kind() string {
	return "latte.NodeChangeEmail"
}

func (n *NodeChangeEmail) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (n *NodeChangeEmail) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return []workflow.Input{
		&InputTakeLoginID{},
	}, nil
}

func (n *NodeChangeEmail) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakeLoginID inputTakeLoginID

	switch {
	case workflow.AsInput(input, &inputTakeLoginID):
		loginID := inputTakeLoginID.GetLoginID()
		spec := &identity.Spec{
			Type: model.IdentityTypeLoginID,
			LoginID: &identity.LoginIDSpec{
				Type:  model.LoginIDKeyTypeEmail,
				Key:   string(model.LoginIDKeyTypeEmail),
				Value: stringutil.NewUserInputString(loginID),
			},
		}

		newInfo, err := deps.Identities.UpdateWithSpec(ctx, n.IdentityBeforeUpdate, spec, identity.NewIdentityOptions{
			LoginIDEmailByPassBlocklistAllowlist: false,
		})
		if err != nil {
			return nil, err
		}

		return workflow.NewNodeSimple(&NodeDoUpdateIdentity{
			IdentityBeforeUpdate: n.IdentityBeforeUpdate,
			IdentityAfterUpdate:  newInfo,
			IsAdminAPI:           false,
		}), nil
	}
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeChangeEmail) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return map[string]interface{}{}, nil
}
