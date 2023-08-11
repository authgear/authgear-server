package workflowconfig

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicIntent(&IntentLoginFlow{})
}

var IntentLoginFlowSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"required": ["login_flow"],
	"properties": {
		"login_flow": { "type": "string" }
	}
}
`)

type IntentLoginFlow struct {
	LoginFlow   string        `json:"login_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
}

var _ workflow.Intent = &IntentLoginFlow{}

func (*IntentLoginFlow) Kind() string {
	return "workflowconfig.IntentLoginFlow"
}

func (*IntentLoginFlow) JSONSchema() *validation.SimpleSchema {
	return IntentLoginFlowSchema
}

func (*IntentLoginFlow) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	// The last node is NodeDoCreateSession.
	// So if MilestoneDoCreateSession is found, this workflow has finished.
	_, ok := FindMilestone[MilestoneDoCreateSession](workflows.Nearest)
	if ok {
		return nil, workflow.ErrEOF
	}
	return nil, nil
}

func (i *IntentLoginFlow) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	switch {
	case len(workflows.Nearest.Nodes) == 0:
		return workflow.NewSubWorkflow(&IntentLoginFlowSteps{
			LoginFlow:   i.LoginFlow,
			JSONPointer: i.JSONPointer,
		}), nil
	case len(workflows.Nearest.Nodes) == 1:
		return workflow.NewNodeSimple(&NodeDoCheckAccountStatus{
			UserID: i.userID(workflows),
		}), nil
	case len(workflows.Nearest.Nodes) == 2:
		return workflow.NewSubWorkflow(&IntentConfirmTerminateOtherSessions{
			UserID: i.userID(workflows),
		}), nil
		// FIXME(workflow): prompt passkey creation
	case len(workflows.Nearest.Nodes) == 3:
		n, err := NewNodeDoCreateSession(ctx, deps, workflows, &NodeDoCreateSession{
			UserID:       i.userID(workflows),
			CreateReason: session.CreateReasonLogin,
			SkipCreate:   workflow.GetSuppressIDPSessionCookie(ctx),
		})
		if err != nil {
			return nil, err
		}
		return workflow.NewNodeSimple(n), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (i *IntentLoginFlow) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return []workflow.Effect{
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			userID := i.userID(workflows)
			usedMethods, err := collectAuthenticationLockoutMethod(ctx, deps, workflows)
			if err != nil {
				return err
			}

			err = deps.Authenticators.ClearLockoutAttempts(userID, usedMethods)
			if err != nil {
				return err
			}

			return nil
		}),
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			// FIXME(workflow): determine isAdminAPI
			isAdminAPI := false
			userID := i.userID(workflows)
			var idpSession *idpsession.IDPSession
			if m, ok := FindMilestone[MilestoneDoCreateSession](workflows.Nearest); ok {
				idpSession, _ = m.MilestoneDoCreateSession()
			}

			// ref: https://github.com/authgear/authgear-server/issues/2930
			// For authentication that involves IDP session will dispatch user.authenticated event here
			// For authentication that suppresses IDP session. e.g. biometric login
			// They are handled in their own node.
			if idpSession == nil {
				return nil
			}

			err := deps.Events.DispatchEvent(&nonblocking.UserAuthenticatedEventPayload{
				UserRef: model.UserRef{
					Meta: model.Meta{
						ID: userID,
					},
				},
				Session:  *idpSession.ToAPIModel(),
				AdminAPI: isAdminAPI,
			})
			if err != nil {
				return err
			}
			return nil
		}),
	}, nil
}

func (*IntentLoginFlow) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (*IntentLoginFlow) userID(workflows workflow.Workflows) string {
	userID, err := getUserID(workflows)
	if err != nil {
		panic(err)
	}

	return userID
}
