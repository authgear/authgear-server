package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterFlow(&IntentReauthFlow{})
}

type IntentReauthFlow struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
}

var _ authflow.PublicFlow = &IntentReauthFlow{}
var _ authflow.EffectGetter = &IntentReauthFlow{}

func (*IntentReauthFlow) Kind() string {
	return "IntentReauthFlow"
}

func (*IntentReauthFlow) FlowType() authflow.FlowType {
	return authflow.FlowTypeReauth
}

func (i *IntentReauthFlow) FlowInit(r authflow.FlowReference, startFrom jsonpointer.T) {
	i.FlowReference = r
}

func (i *IntentReauthFlow) FlowFlowReference() authflow.FlowReference {
	return i.FlowReference
}

func (i *IntentReauthFlow) FlowRootObject(deps *authflow.Dependencies) (config.AuthenticationFlowObject, error) {
	return flowRootObject(deps, i.FlowReference)
}

func (i *IntentReauthFlow) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// The last node is NodeDidReauthenticate.
	// So if MilestoneDidReauthenticate is found, this flow has finished.
	_, ok := authflow.FindMilestone[MilestoneDidReauthenticate](flows.Nearest)
	if ok {
		return nil, authflow.ErrEOF
	}
	return nil, nil
}

func (i *IntentReauthFlow) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, _ authflow.Input) (*authflow.Node, error) {
	switch {
	case len(flows.Nearest.Nodes) == 0:
		return authflow.NewSubFlow(&IntentReauthFlowSteps{
			FlowReference: i.FlowReference,
			JSONPointer:   i.JSONPointer,
		}), nil
	case len(flows.Nearest.Nodes) == 1:
		n, err := NewNodeDidReauthenticate(ctx, deps, flows, &NodeDidReauthenticate{
			UserID: i.userID(flows),
		})
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(n), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentReauthFlow) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]authflow.Effect, error) {
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

		// Reauth does not create new session.
		// So we do not dispatch user.authenticated here.
	}, nil
}

func (*IntentReauthFlow) userID(flows authflow.Flows) string {
	userID, err := getUserID(flows)
	if err != nil {
		panic(err)
	}

	return userID
}
