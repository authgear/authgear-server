package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicIntent(&IntentAuthenticate{})
}

var IntentAuthenticateSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false
	}
`)

type IntentAuthenticate struct{}

func (*IntentAuthenticate) Kind() string {
	return "latte.IntentAuthenticate"
}

func (*IntentAuthenticate) JSONSchema() *validation.SimpleSchema {
	return IntentAuthenticateSchema
}

func (*IntentAuthenticate) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	if len(w.Nodes) == 0 {
		return []workflow.Input{
			&InputTakeLoginID{},
		}, nil
	}
	// IntentAuthenticate has only one node, which could be IntentSignup or IntentLogin.
	return nil, workflow.ErrEOF
}

func (*IntentAuthenticate) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
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
		exactMatch, _, err := deps.Identities.SearchBySpec(spec)
		if err != nil {
			return nil, err
		}
		if exactMatch == nil {
			return workflow.NewSubWorkflow(&IntentSignup{}), nil
		}
		return workflow.NewSubWorkflow(&IntentLogin{}), nil
	default:
		return nil, workflow.ErrIncompatibleInput
	}
}

func (*IntentAuthenticate) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	// IntentAuthenticate has no effects.
	// The effects would be in IntentSignup or IntentLogin.
	return nil, nil
}

func (*IntentAuthenticate) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}
