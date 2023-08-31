package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicIntent(&IntentChangePassword{})
}

var IntentChangePasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false
	}
`)

type IntentChangePassword struct {
}

func (*IntentChangePassword) Kind() string {
	return "latte.IntentChangePassword"
}

func (*IntentChangePassword) JSONSchema() *validation.SimpleSchema {
	return IntentChangePasswordSchema
}

func (*IntentChangePassword) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	switch len(workflows.Nearest.Nodes) {
	case 0:
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentChangePassword) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	userID := session.GetUserID(ctx)
	if userID == nil {
		return nil, apierrors.NewUnauthorized("authentication required")
	}

	return workflow.NewNodeSimple(&NodeChangePassword{
		UserID:            *userID,
		AuthenticatorKind: authenticator.KindPrimary,
	}), nil
}

func (i *IntentChangePassword) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentChangePassword) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
