package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var loginFlowCheckAccountStatusLogger = slogutil.NewLogger("login-flow-check-account-status")

func init() {
	authflow.RegisterIntent(&IntentLoginFlowStepCheckAccountStatus{})
}

type IntentLoginFlowStepCheckAccountStatus struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	StepName      string                 `json:"step_name,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
}

var _ authflow.Intent = &IntentLoginFlowStepCheckAccountStatus{}

func (*IntentLoginFlowStepCheckAccountStatus) Kind() string {
	return "IntentLoginFlowStepCheckAccountStatus"
}

func (i *IntentLoginFlowStepCheckAccountStatus) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// This step does not require any input.
	if len(flows.Nearest.Nodes) == 0 {
		return nil, nil
	}

	return nil, authflow.ErrEOF
}

func (i *IntentLoginFlowStepCheckAccountStatus) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	u, err := deps.Users.GetRaw(ctx, i.UserID)
	if err != nil {
		return nil, err
	}

	err = u.AccountStatus().Check()
	if err != nil {
		apiError := apierrors.AsAPIError(err)
		if apiError != nil {
			dispatchErr := deps.Events.DispatchEventImmediately(ctx, &nonblocking.AuthenticationBlockedEventPayload{
				UserRef: *u.ToRef(),
				Error:   apiError,
			})
			if dispatchErr != nil {
				loginFlowCheckAccountStatusLogger.GetLogger(ctx).WithError(err).Error(ctx, "failed to dispatch event")
			}
		}
		return nil, err
	}

	return authflow.NewNodeSimple(&NodeSentinel{}), nil
}
