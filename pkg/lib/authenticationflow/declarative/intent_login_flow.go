package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
)

func init() {
	authflow.RegisterFlow(&IntentLoginFlow{})
}

type IntentLoginFlow struct {
	LoginFlow   string        `json:"login_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
}

var _ authflow.PublicFlow = &IntentLoginFlow{}
var _ authflow.EffectGetter = &IntentLoginFlow{}

func (*IntentLoginFlow) Kind() string {
	return "IntentLoginFlow"
}

func (*IntentLoginFlow) FlowType() authflow.FlowType {
	return authflow.FlowTypeLogin
}

func (i *IntentLoginFlow) FlowInit(r authflow.FlowReference) {
	i.LoginFlow = r.ID
}

func (*IntentLoginFlow) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// The last node is NodeDoCreateSession.
	// So if MilestoneDoCreateSession is found, this flow has finished.
	_, ok := authflow.FindMilestone[MilestoneDoCreateSession](flows.Nearest)
	if ok {
		return nil, authflow.ErrEOF
	}
	return nil, nil
}

func (i *IntentLoginFlow) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	switch {
	case len(flows.Nearest.Nodes) == 0:
		return authflow.NewSubFlow(&IntentLoginFlowSteps{
			LoginFlow:   i.LoginFlow,
			JSONPointer: i.JSONPointer,
		}), nil
	case len(flows.Nearest.Nodes) == 1:
		return authflow.NewNodeSimple(&NodeDoCheckAccountStatus{
			UserID: i.userID(flows),
		}), nil
	case len(flows.Nearest.Nodes) == 2:
		return authflow.NewSubFlow(&IntentConfirmTerminateOtherSessions{
			UserID: i.userID(flows),
		}), nil
		// FIXME(authflow): prompt passkey creation
	case len(flows.Nearest.Nodes) == 3:
		n, err := NewNodeDoCreateSession(ctx, deps, flows, &NodeDoCreateSession{
			UserID:       i.userID(flows),
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
			userID := i.userID(flows)
			usedMethods, err := collectAuthenticationLockoutMethod(ctx, deps, flows)
			if err != nil {
				return err
			}

			err = deps.Authenticators.ClearLockoutAttempts(userID, usedMethods)
			if err != nil {
				return err
			}

			return nil
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			// FIXME(authflow): determine isAdminAPI
			isAdminAPI := false
			userID := i.userID(flows)
			var idpSession *idpsession.IDPSession
			if m, ok := authflow.FindMilestone[MilestoneDoCreateSession](flows.Nearest); ok {
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

func (*IntentLoginFlow) userID(flows authflow.Flows) string {
	userID, err := getUserID(flows)
	if err != nil {
		panic(err)
	}

	return userID
}
