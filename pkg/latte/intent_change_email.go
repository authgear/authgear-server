package latte

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/stringutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicIntent(&IntentChangeEmail{})
}

var IntentChangeEmailSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false
	}
`)

type IntentChangeEmail struct {
}

func (*IntentChangeEmail) Kind() string {
	return "latte.IntentChangeEmail"
}

func (*IntentChangeEmail) JSONSchema() *validation.SimpleSchema {
	return IntentChangeEmailSchema
}

func (*IntentChangeEmail) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	switch len(workflows.Nearest.Nodes) {
	case 0:
		return []workflow.Input{
			&InputTakeCurrentLoginID{},
		}, nil
	case 1:
		return nil, nil
	case 2:
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentChangeEmail) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	userID := session.GetUserID(ctx)
	if userID == nil {
		return nil, apierrors.NewUnauthorized("authentication required")
	}

	switch len(workflows.Nearest.Nodes) {
	case 0:
		var inputTakeCurrentLoginID inputTakeCurrentLoginID
		switch {
		case workflow.AsInput(input, &inputTakeCurrentLoginID):
			loginID := inputTakeCurrentLoginID.GetCurrentLoginID()
			spec := &identity.Spec{
				Type: model.IdentityTypeLoginID,
				LoginID: &identity.LoginIDSpec{
					Type:  model.LoginIDKeyTypeEmail,
					Key:   string(model.LoginIDKeyTypeEmail),
					Value: stringutil.NewUserInputString(loginID),
				},
			}
			exactMatch, _, err := deps.Identities.SearchBySpec(ctx, spec)
			if err != nil {
				return nil, err
			}
			if exactMatch == nil || exactMatch.UserID != *userID {
				return nil, api.ErrIdentityNotFound
			}

			return workflow.NewNodeSimple(&NodeChangeEmail{
				UserID:               *userID,
				IdentityBeforeUpdate: exactMatch,
			}), nil
		}
	case 1:
		iden := i.newIdentityInfo(workflows.Nearest)
		return workflow.NewNodeSimple(&NodePopulateStandardAttributes{
			Identity: iden,
		}), nil
	case 2:
		iden := i.newIdentityInfo(workflows.Nearest)
		return workflow.NewSubWorkflow(&IntentVerifyIdentity{
			Identity:     iden,
			IsFromSignUp: false,
		}), nil
		// We do not need to manually create a new authenticator because update identity will handle that
		// This feature is added in https://github.com/authgear/2023C01-latte-customization/issues/212
	}
	return nil, workflow.ErrIncompatibleInput
}

func (i *IntentChangeEmail) newIdentityInfo(w *workflow.Workflow) *identity.Info {
	node, ok := workflow.FindSingleNode[*NodeDoUpdateIdentity](w)
	if !ok {
		panic(fmt.Errorf("workflow: expected NodeDoUpdateIdentity"))
	}

	return node.IdentityAfterUpdate
}

func (i *IntentChangeEmail) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentChangeEmail) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
