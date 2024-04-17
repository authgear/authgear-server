package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

func init() {
	authflow.RegisterFlow(&IntentPromoteFlow{})
}

type IntentPromoteFlow struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
}

var _ authflow.PublicFlow = &IntentPromoteFlow{}
var _ authflow.EffectGetter = &IntentPromoteFlow{}

func (*IntentPromoteFlow) Kind() string {
	return "IntentPromoteFlow"
}

func (*IntentPromoteFlow) FlowType() authflow.FlowType {
	return authflow.FlowTypePromote
}
func (i *IntentPromoteFlow) FlowInit(r authflow.FlowReference, startFrom jsonpointer.T) {
	i.FlowReference = r
}

func (i *IntentPromoteFlow) FlowFlowReference() authflow.FlowReference {
	return i.FlowReference
}

func (i *IntentPromoteFlow) FlowRootObject(deps *authflow.Dependencies) (config.AuthenticationFlowObject, error) {
	return flowRootObject(deps, i.FlowReference)
}

func (i *IntentPromoteFlow) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// The list of nodes looks like
	// 1 NodeDoUseAnonymousUser
	// 1 IntentPromoteFlowSteps
	// 1 NodeDoCreateSession
	// So if MilestoneDoCreateSession is found, this flow has finished.
	_, ok := authflow.FindMilestone[MilestoneDoCreateSession](flows.Nearest)
	if ok {
		return nil, authflow.ErrEOF
	}
	return nil, nil
}

func (i *IntentPromoteFlow) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, _ authflow.Input) (*authflow.Node, error) {
	switch {
	case len(flows.Nearest.Nodes) == 0:
		node, err := NewNodeDoUseAnonymousUser(ctx, deps)
		if err != nil {
			return nil, err
		}
		return authflow.NewNodeSimple(node), nil
	case len(flows.Nearest.Nodes) == 1:
		return authflow.NewSubFlow(&IntentPromoteFlowSteps{
			FlowReference: i.FlowReference,
			JSONPointer:   i.JSONPointer,
			UserID:        i.userID(flows.Nearest),
		}), nil
	case len(flows.Nearest.Nodes) == 2:
		n, err := NewNodeDoCreateSession(ctx, deps, flows, &NodeDoCreateSession{
			UserID:       i.userID(flows.Nearest),
			CreateReason: session.CreateReasonPromote,
			SkipCreate:   authflow.GetSuppressIDPSessionCookie(ctx),
		})
		if err != nil {
			return nil, err
		}
		return authflow.NewNodeSimple(n), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentPromoteFlow) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	return []authflow.Effect{
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			isAnonymous := true
			// Apply rate limit on sign up.
			spec := SignupPerIPRateLimitBucketSpec(deps.Config.Authentication, isAnonymous, string(deps.RemoteIP))
			err := deps.RateLimiter.Allow(spec)
			if err != nil {
				return err
			}
			return nil
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			// Remove the anonymous identity
			anonymousIden := i.anonymousIdentity(flows.Nearest)

			err := deps.Identities.Delete(anonymousIden)
			if err != nil {
				return err
			}

			return nil
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			userID := i.userID(flows.Nearest)
			anonUserRef := model.UserRef{
				Meta: model.Meta{
					ID: userID,
				},
			}

			// We have remove the anonymous identity in the previous effect,
			// so we can simply list the identities here.
			identities, err := deps.Identities.ListByUser(userID)
			if err != nil {
				return err
			}

			var identityModels []model.Identity
			for _, info := range identities {
				identityModels = append(identityModels, info.ToModel())
			}

			isAdminAPI := false
			err = deps.Events.DispatchEventOnCommit(&nonblocking.UserAnonymousPromotedEventPayload{
				AnonymousUserRef: anonUserRef,
				UserRef:          anonUserRef,
				Identities:       identityModels,
				AdminAPI:         isAdminAPI,
			})
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}

func (i *IntentPromoteFlow) anonymousIdentity(f *authflow.Flow) *identity.Info {
	m, ok := authflow.FindMilestone[MilestoneDoUseAnonymousUser](f)
	if !ok {
		panic(fmt.Errorf("expected userID"))
	}

	info := m.MilestoneDoUseAnonymousUser()
	return info
}

func (i *IntentPromoteFlow) userID(f *authflow.Flow) string {
	return i.anonymousIdentity(f).UserID
}
