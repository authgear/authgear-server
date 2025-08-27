package declarative

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

var signupFlowLogger = slogutil.NewLogger("signup-flow")

func init() {
	authflow.RegisterFlow(&IntentSignupFlow{})
}

type IntentSignupFlow struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
}

var _ authflow.PublicFlow = &IntentSignupFlow{}
var _ authflow.EffectGetter = &IntentSignupFlow{}
var _ authflow.Milestone = &IntentSignupFlow{}
var _ MilestoneSwitchToExistingUser = &IntentSignupFlow{}

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
	return getFlowRootObject(deps, i.FlowReference)
}

func (*IntentSignupFlow) Milestone() {}
func (i *IntentSignupFlow) MilestoneSwitchToExistingUser(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, newUserID string) error {
	milestone, _, ok := authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateUser](flows)
	if ok {
		milestone.MilestoneDoCreateUserUseExisting(newUserID)
	}
	return nil
}

func (i *IntentSignupFlow) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// The list of nodes looks like
	// 1 NodeDoCreateUser
	// 1 IntentSignupFlowSteps
	// 1 NodeDoCreateSession
	// So if MarkerDoCreateSession is found, this flow has finished.
	_, _, ok := authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateSession](flows)
	if ok {
		return nil, authflow.ErrEOF
	}
	return nil, nil
}

func (i *IntentSignupFlow) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	if deps.Config.Authentication.PublicSignupDisabled {
		return nil, ErrNoPublicSignup
	}

	switch {
	case len(flows.Nearest.Nodes) == 0:
		return NewNodePreInitialize(ctx, deps, flows)
	case len(flows.Nearest.Nodes) == 1:
		return authflow.NewNodeSimple(&NodeDoCreateUser{
			UserID: uuid.New(),
		}), nil
	case len(flows.Nearest.Nodes) == 2:
		userID, _ := i.userID(flows)
		if userID == "" {
			panic(fmt.Errorf("expected userID to be non empty in IntentSignupFlow"))
		}
		return authflow.NewSubFlow(&IntentSignupFlowSteps{
			FlowReference: i.FlowReference,
			JSONPointer:   i.JSONPointer,
			UserID:        userID,
		}), nil
	case len(flows.Nearest.Nodes) == 3:
		userID, createUser := i.userID(flows)
		n, err := NewNodeDoCreateSession(ctx, deps, flows, &NodeDoCreateSession{
			UserID:       userID,
			CreateReason: session.CreateReasonSignup,
			SkipCreate:   !createUser || authflow.GetSuppressIDPSessionCookie(ctx),
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
			specs := ratelimit.RateLimitAuthenticationSignup.ResolveBucketSpecs(deps.Config, deps.FeatureConfig, deps.RateLimitsEnvConfig, &ratelimit.ResolveBucketSpecOptions{
				IPAddress: string(deps.RemoteIP),
			})
			for _, spec := range specs {
				spec := *spec
				failed, err := deps.RateLimiter.Allow(ctx, spec)
				if err != nil {
					return err
				}
				if err := failed.Error(); err != nil {
					return err
				}
			}
			return nil
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			userID, createUser := i.userID(flows)
			if !createUser {
				return nil
			}

			now := deps.Clock.NowUTC()
			// NOTE(DEV-2982): This is for debugging the session lost problem
			logger := signupFlowLogger.GetLogger(ctx)
			logger.WithSkipLogging().Error(ctx, "updated last login",
				slog.String("user_id", userID))
			return deps.Users.UpdateLoginTime(ctx, userID, now)
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			userID, createUser := i.userID(flows)
			if !createUser {
				// The creation is skipped for some reason, such as entered account linking flow
				return nil
			}
			isAdminAPI := false

			u, err := deps.Users.GetRaw(ctx, userID)
			if err != nil {
				return err
			}

			identities, err := deps.Identities.ListByUser(ctx, userID)
			if err != nil {
				return err
			}

			authenticators, err := deps.Authenticators.List(ctx, userID)
			if err != nil {
				return err
			}

			err = deps.Users.AfterCreate(ctx,
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

func (i *IntentSignupFlow) userID(flows authflow.Flows) (userID string, createUser bool) {
	m, _, ok := authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateUser](flows)
	if !ok {
		panic(fmt.Errorf("expected userID"))
	}

	return m.MilestoneDoCreateUser()
}
