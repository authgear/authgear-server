package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func init() {
	authflow.RegisterFlow(&IntentSignupFlow{})
}

type IntentSignupFlow struct {
	SignupFlow  string        `json:"signup_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
}

var _ authflow.PublicFlow = &IntentSignupFlow{}
var _ authflow.EffectGetter = &IntentSignupFlow{}

func (*IntentSignupFlow) Kind() string {
	return "IntentSignupFlow"
}

func (*IntentSignupFlow) FlowType() authflow.FlowType {
	return authflow.FlowTypeSignup
}

func (i *IntentSignupFlow) FlowInit(r authflow.FlowReference) {
	i.SignupFlow = r.ID
}

func (i *IntentSignupFlow) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// The list of nodes looks like
	// 1 NodeDoCreateUser
	// 1 IntentSignupFlowSteps
	// 1 NodeDoCreateSession
	// So if MarkerDoCreateSession is found, this flow has finished.
	_, ok := authflow.FindMilestone[MilestoneDoCreateSession](flows.Nearest)
	if ok {
		return nil, authflow.ErrEOF
	}
	return nil, nil
}

func (i *IntentSignupFlow) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, _ authflow.Input) (*authflow.Node, error) {
	switch {
	case len(flows.Nearest.Nodes) == 0:
		return authflow.NewNodeSimple(&NodeDoCreateUser{
			UserID: uuid.New(),
		}), nil
	case len(flows.Nearest.Nodes) == 1:
		return authflow.NewSubFlow(&IntentSignupFlowSteps{
			SignupFlow:  i.SignupFlow,
			JSONPointer: i.JSONPointer,
			UserID:      i.userID(flows.Nearest),
		}), nil
	case len(flows.Nearest.Nodes) == 2:
		n, err := NewNodeDoCreateSession(ctx, deps, flows, &NodeDoCreateSession{
			UserID:       i.userID(flows.Nearest),
			CreateReason: session.CreateReasonSignup,
			SkipCreate:   authflow.GetSuppressIDPSessionCookie(ctx),
		})
		if err != nil {
			return nil, err
		}
		return authflow.NewNodeSimple(n), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentSignupFlow) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	return []authflow.Effect{
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			// Apply rate limit on sign up.
			spec := SignupPerIPRateLimitBucketSpec(deps.Config.Authentication, false, string(deps.RemoteIP))
			err := deps.RateLimiter.Allow(spec)
			if err != nil {
				return err
			}
			return nil
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			userID := i.userID(flows.Nearest)
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

func (i *IntentSignupFlow) userID(w *authflow.Flow) string {
	m, ok := authflow.FindMilestone[MilestoneDoCreateUser](w)
	if !ok {
		panic(fmt.Errorf("expected userID"))
	}

	id := m.MilestoneDoCreateUser()

	return id
}
