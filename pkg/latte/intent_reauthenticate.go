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
	workflow.RegisterPublicIntent(&IntentReauthenticate{})
}

var IntentReauthenticateSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false
	}
`)

type IntentReauthenticate struct {
}

func (*IntentReauthenticate) Kind() string {
	return "latte.IntentReauthenticate"
}

func (*IntentReauthenticate) JSONSchema() *validation.SimpleSchema {
	return IntentReauthenticateSchema
}

func (*IntentReauthenticate) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	switch len(w.Nodes) {
	case 0:
		return nil, nil
	case 1:
		return nil, nil
	}

	return nil, workflow.ErrEOF
}

func (i *IntentReauthenticate) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	userID, err := i.userID(ctx)
	if err != nil {
		return nil, err
	}

	switch len(w.Nodes) {
	case 0:
		return workflow.NewSubWorkflow(&IntentAuthenticatePassword{
			UserID:            userID,
			AuthenticatorKind: authenticator.KindPrimary,
		}), nil
	case 1:
		mode := EnsureSessionModeUpdateOrCreate
		if workflow.GetSuppressIDPSessionCookie(ctx) {
			mode = EnsureSessionModeNoop
		}
		return workflow.NewSubWorkflow(&IntentEnsureSession{
			UserID:       userID,
			CreateReason: session.CreateReasonReauthenticate,
			AMR:          GetAMR(w),
			Mode:         mode,
		}), nil
	}

	return nil, workflow.ErrIncompatibleInput

}

func (*IntentReauthenticate) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentReauthenticate) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}

func (*IntentReauthenticate) userID(ctx context.Context) (string, error) {
	userID := workflow.GetUserIDHint(ctx)
	if userID == "" {
		return "", apierrors.NewInvalid("this workflow must be triggered in a reauthentication session")
	}
	return userID, nil
}
