package workflowconfig

import (
	"context"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterIntent(&IntentInspectDeviceToken{})
}

type IntentInspectDeviceToken struct {
	UserID string `json:"user_id,omitempty"`
}

var _ workflow.Intent = &IntentInspectDeviceToken{}
var _ workflow.Milestone = &IntentInspectDeviceToken{}
var _ MilestoneDeviceTokenInspected = &IntentInspectDeviceToken{}

func (*IntentInspectDeviceToken) Kind() string {
	return "workflowconfig.IntentInspectDeviceToken"
}

func (*IntentInspectDeviceToken) Milestone()                     {}
func (*IntentInspectDeviceToken) MilestoneDeviceTokenInspected() {}

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
