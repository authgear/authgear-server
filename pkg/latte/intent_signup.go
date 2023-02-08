package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/uuid"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentSignup{})
}

var IntentSignupSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false
	}
`)

type IntentSignup struct{}

func (*IntentSignup) Kind() string {
	return "latte.IntentSignup"
}

func (*IntentSignup) JSONSchema() *validation.SimpleSchema {
	return IntentSignupSchema
}

func (*IntentSignup) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	switch len(w.Nodes) {
	case 0:
		return nil, nil
	// case 1:
	// IntentCreateLoginID (Phone + OTP)
	// case 2:
	// IntentCreateLoginID (Email, no verification)
	// case 3:
	// 4. IntentCreatePrimaryPassword
	default:
		return nil, workflow.ErrEOF
	}
}

func (*IntentSignup) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	switch len(w.Nodes) {
	case 0:
		return workflow.NewNodeSimple(&NodeDoCreateUser{
			UserID: uuid.New(),
		}), nil
	default:
		return nil, workflow.ErrIncompatibleInput
	}
}

func (*IntentSignup) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	// TODO(workflow): Fire user.created.
	return nil, nil
}

func (*IntentSignup) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}
