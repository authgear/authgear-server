package declarative

import (
	"context"
	"log/slog"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var loginFlowLogger = slogutil.NewLogger("login-flow")

func init() {
	authflow.RegisterFlow(&IntentLoginFlow{})
}

type IntentLoginFlow struct {
	TargetUserID  string                 `json:"target_user_id,omitempty"`
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
}

var _ authflow.PublicFlow = &IntentLoginFlow{}
var _ authflow.EffectGetter = &IntentLoginFlow{}

func (*IntentLoginFlow) Kind() string {
	return "IntentLoginFlow"
}

func (*IntentLoginFlow) FlowType() authflow.FlowType {
	return authflow.FlowTypeLogin
}

func (i *IntentLoginFlow) FlowInit(r authflow.FlowReference, startFrom jsonpointer.T) {
	i.FlowReference = r
}

func (i *IntentLoginFlow) FlowFlowReference() authflow.FlowReference {
	return i.FlowReference
}

func (i *IntentLoginFlow) FlowRootObject(deps *authflow.Dependencies) (config.AuthenticationFlowObject, error) {
	return getFlowRootObject(deps, i.FlowReference)
}

func (*IntentLoginFlow) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// The last node is NodeDoCreateSession.
	// So if MilestoneDoCreateSession is found, this flow has finished.
	_, _, ok := authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateSession](flows)
	if ok {
		return nil, authflow.ErrEOF
	}
	return nil, nil
}

func (i *IntentLoginFlow) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	switch {
	case len(flows.Nearest.Nodes) == 0:
		return NewNodePreInitialize(ctx, deps, flows)
	case len(flows.Nearest.Nodes) == 1:
		return authflow.NewSubFlow(&IntentLoginFlowSteps{
			FlowReference: i.FlowReference,
			JSONPointer:   i.JSONPointer,
		}), nil
	case len(flows.Nearest.Nodes) == 2:
		userID, err := i.userID(flows)
		if err != nil {
			return nil, err
		}
		n, err := NewNodeDoCreateSession(ctx, deps, flows, &NodeDoCreateSession{
			UserID:       userID,
			CreateReason: session.CreateReasonLogin,
			SkipCreate:   authflow.GetSuppressIDPSessionCookie(ctx),
		})
		if err != nil {
			return nil, err
		}
		return authflow.NewNodeSimple(n), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentLoginFlow) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]authflow.Effect, error) {
	return []authflow.Effect{
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			userID, err := i.userID(flows)
			if err != nil {
				return err
			}
			now := deps.Clock.NowUTC()
			// NOTE(DEV-2982): This is for debugging the session lost problem
			logger := loginFlowLogger.GetLogger(ctx)
			logger.WithSkipLogging().Error(ctx, "updated last login",
				slog.String("user_id", userID))
			return deps.Users.UpdateLoginTime(ctx, userID, now)
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			userID, err := i.userID(flows)
			if err != nil {
				return err
			}
			usedMethods, err := collectAuthenticationLockoutMethod(ctx, deps, flows)
			if err != nil {
				return err
			}

			err = deps.Authenticators.ClearLockoutAttempts(ctx, userID, usedMethods)
			if err != nil {
				return err
			}

			return nil
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			// FIXME(authflow): determine isAdminAPI
			isAdminAPI := false
			userID, err := i.userID(flows)
			if err != nil {
				return err
			}
			var idpSession *idpsession.IDPSession
			if m, _, ok := authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateSession](flows); ok {
				idpSession, _ = m.MilestoneDoCreateSession()
			}

			// ref: https://github.com/authgear/authgear-server/issues/2930
			// For authentication that involves IDP session will dispatch user.authenticated event here
			// For authentication that suppresses IDP session. e.g. biometric login
			// They are handled in their own node.
			if idpSession == nil {
				// NOTE(DEV-2982): This is for debugging the session lost problem
				logger := loginFlowLogger.GetLogger(ctx)
				logger.WithSkipLogging().Error(ctx, "user.authenticated event skipped because IDP session is nil",
					slog.String("user_id", userID))
				return nil
			}

			err = deps.Events.DispatchEventOnCommit(ctx, &nonblocking.UserAuthenticatedEventPayload{
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

func (i *IntentLoginFlow) userID(flows authflow.Flows) (string, error) {
	userID, err := getUserID(flows)
	if err != nil {
		panic(err)
	}

	if i.TargetUserID != "" && i.TargetUserID != userID {
		return "", ErrDifferentUserID
	}

	return userID, nil
}
