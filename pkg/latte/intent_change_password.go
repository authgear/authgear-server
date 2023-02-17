package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicIntent(&IntentChangePassword{})
}

var IntentChangePasswordSchema = validation.NewSimpleSchema(`{}`)

type IntentChangePassword struct {
}

func (*IntentChangePassword) Kind() string {
	return "latte.IntentChangePassword"
}

func (*IntentChangePassword) JSONSchema() *validation.SimpleSchema {
	return IntentChangePasswordSchema
}

func (*IntentChangePassword) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	if len(w.Nodes) == 0 {
		return []workflow.Input{
			&InputTakeLoginID{},
		}, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentChangePassword) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	var inputTakeLoginID inputTakeLoginID

	switch {
	case workflow.AsInput(input, &inputTakeLoginID):
		loginID := inputTakeLoginID.GetLoginID()
		spec := &identity.Spec{
			Type: model.IdentityTypeLoginID,
			LoginID: &identity.LoginIDSpec{
				Value: loginID,
			},
		}
		identity, _, err := deps.Identities.SearchBySpec(spec)
		if err != nil {
			return nil, err
		}
		if identity == nil {
			return nil, workflow.ErrIncompatibleInput
		}

		return workflow.NewNodeSimple(&NodeChangePassword{
			UserID:            identity.LoginID.UserID,
			AuthenticatorKind: authenticator.KindPrimary,
		}), nil
	default:
		return nil, workflow.ErrIncompatibleInput
	}
}

func (i *IntentChangePassword) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentChangePassword) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}
