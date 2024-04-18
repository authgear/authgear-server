package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func init() {
	authflow.RegisterFlow(&IntentSignupFlow{})
}

type IntentSignupFlow struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
}

var _ authflow.PublicFlow = &IntentSignupFlow{}
var _ authflow.EffectGetter = &IntentSignupFlow{}

func (*IntentSignupFlow) Kind() string {
	return "IntentSignupFlow"
}

func (*IntentSignupFlow) FlowType() authflow.FlowType {
	return authflow.FlowTypeSignup
}

func (i *IntentSignupFlow) FlowInit(r authflow.FlowReference, startFrom jsonpointer.T) {
	i.FlowReference = r
}

func (i *IntentSignupFlow) FlowFlowReference() authflow.FlowReference {
	return i.FlowReference
}

func (i *IntentSignupFlow) FlowRootObject(deps *authflow.Dependencies) (config.AuthenticationFlowObject, error) {
	return flowRootObject(deps, i.FlowReference)
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
	if deps.Config.Authentication.PublicSignupDisabled {
		return nil, ErrNoPublicSignup
	}

	switch {
	case len(flows.Nearest.Nodes) == 0:
		return authflow.NewNodeSimple(&NodeDoCreateUser{
			UserID: uuid.New(),
		}), nil
	case len(flows.Nearest.Nodes) == 1:
		return authflow.NewSubFlow(&IntentSignupFlowSteps{
			FlowReference: i.FlowReference,
			JSONPointer:   i.JSONPointer,
			UserID:        i.userID(flows.Nearest),
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
			if userID == "" {
				// The creation is skipped for some reason, such as entered account linking flow
				return nil
			}
			isAdminAPI := false

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
