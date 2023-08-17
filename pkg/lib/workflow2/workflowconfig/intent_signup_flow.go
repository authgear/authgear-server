package workflowconfig

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func init() {
	workflow.RegisterFlow(&IntentSignupFlow{})
}

type IntentSignupFlow struct {
	SignupFlow  string        `json:"signup_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
}

var _ workflow.Flow = &IntentSignupFlow{}
var _ workflow.EffectGetter = &IntentSignupFlow{}

func (*IntentSignupFlow) Kind() string {
	return "workflowconfig.IntentSignupFlow"
}

func (*IntentSignupFlow) FlowType() workflow.FlowType {
	return workflow.FlowTypeSignup
}

func (i *IntentSignupFlow) FlowInit(r workflow.FlowReference) {
	i.SignupFlow = r.ID
}

func (i *IntentSignupFlow) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	// The list of nodes looks like
	// 1 NodeDoCreateUser
	// 1 IntentSignupFlowSteps
	// 1 NodeDoCreateSession
	// So if MarkerDoCreateSession is found, this workflow has finished.
	_, ok := workflow.FindMilestone[MilestoneDoCreateSession](workflows.Nearest)
	if ok {
		return nil, workflow.ErrEOF
	}
	return nil, nil
}

func (i *IntentSignupFlow) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	switch {
	case len(workflows.Nearest.Nodes) == 0:
		return workflow.NewNodeSimple(&NodeDoCreateUser{
			UserID: uuid.New(),
		}), nil
	case len(workflows.Nearest.Nodes) == 1:
		return workflow.NewSubWorkflow(&IntentSignupFlowSteps{
			SignupFlow:  i.SignupFlow,
			JSONPointer: i.JSONPointer,
			UserID:      i.userID(workflows.Nearest),
		}), nil
	case len(workflows.Nearest.Nodes) == 2:
		n, err := NewNodeDoCreateSession(ctx, deps, workflows, &NodeDoCreateSession{
			UserID:       i.userID(workflows.Nearest),
			CreateReason: session.CreateReasonSignup,
			SkipCreate:   workflow.GetSuppressIDPSessionCookie(ctx),
		})
		if err != nil {
			return nil, err
		}
		return workflow.NewNodeSimple(n), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (i *IntentSignupFlow) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			// Apply rate limit on sign up.
			spec := SignupPerIPRateLimitBucketSpec(deps.Config.Authentication, false, string(deps.RemoteIP))
			err := deps.RateLimiter.Allow(spec)
			if err != nil {
				return err
			}
			return nil
		}),
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			userID := i.userID(workflows.Nearest)
			isAdminAPI := false
			uiParam := uiparam.GetUIParam(ctx)

			u, err := deps.Users.GetRaw(userID)
			if err != nil {
				return err
			}

			identities, err := deps.Identities.ListByUser(userID)
			if err != nil {
				return err
			}

			authenticators, err := deps.Authenticators.List(userID)
			if err != nil {
				return err
			}

			err = deps.Users.AfterCreate(
				u,
				identities,
				authenticators,
				isAdminAPI,
				uiParam,
			)
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}

func (i *IntentSignupFlow) userID(w *workflow.Workflow) string {
	m, ok := workflow.FindMilestone[MilestoneDoCreateUser](w)
	if !ok {
		panic(fmt.Errorf("expected userID"))
	}

	id := m.MilestoneDoCreateUser()

	return id
}
