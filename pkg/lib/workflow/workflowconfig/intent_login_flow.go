package workflowconfig

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/session"
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
		// FIXME(workflow): check account status
	case len(workflows.Nearest.Nodes) == 1:
		// FIXME(workflow): create session
		node := NewNodeDoCreateSession(deps, &NodeDoCreateSession{
			UserID:       i.userID(workflows),
			CreateReason: session.CreateReasonLogin,
			SkipCreate:   workflow.GetSuppressIDPSessionCookie(ctx),
		})
		return workflow.NewNodeSimple(node), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (i *IntentLoginFlow) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	// FIXME(workflow): login effects.
	return nil, nil
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
