package workflowconfig

import (
	"context"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentInspectDeviceToken{})
}

type IntentInspectDeviceToken struct {
	UserID string `json:"user_id,omitempty"`
}

var _ Milestone = &IntentInspectDeviceToken{}

func (*IntentInspectDeviceToken) Milestone() {}

var _ MilestoneDeviceTokenInspected = &IntentInspectDeviceToken{}

func (*IntentInspectDeviceToken) MilestoneDeviceTokenInspected() {}

var IntentInspectDeviceTokenSchema = validation.NewSimpleSchema(`{}`)

var _ workflow.Intent = &IntentInspectDeviceToken{}

func (*IntentInspectDeviceToken) Kind() string {
	return "workflowconfig.IntentInspectDeviceToken"
}

func (*IntentInspectDeviceToken) JSONSchema() *validation.SimpleSchema {
	return IntentInspectDeviceTokenSchema
}

func (*IntentInspectDeviceToken) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		return nil, nil
	}

	return nil, workflow.ErrEOF
}

func (i *IntentInspectDeviceToken) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		deviceTokenCookie, err := deps.Cookies.GetCookie(deps.HTTPRequest, deps.MFADeviceTokenCookie.Def)
		// End this workflow if there is no cookie.
		if errors.Is(err, http.ErrNoCookie) {
			return workflow.NewNodeSimple(&NodeSentinel{}), nil
		} else if err != nil {
			return nil, err
		}

		deviceToken := deviceTokenCookie.Value

		err = deps.MFA.VerifyDeviceToken(i.UserID, deviceToken)
		if errors.Is(err, mfa.ErrDeviceTokenNotFound) {
			deviceTokenCookie = deps.Cookies.ClearCookie(deps.MFADeviceTokenCookie.Def)
			return workflow.NewNodeSimple(&NodeDoClearDeviceTokenCookie{
				Cookie: deviceTokenCookie,
			}), nil
		} else if err != nil {
			return nil, err
		}

		return workflow.NewNodeSimple(&NodeDoUseDeviceToken{}), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*IntentInspectDeviceToken) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return nil, nil
}

func (*IntentInspectDeviceToken) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
