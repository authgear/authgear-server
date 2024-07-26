package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func init() {
	authflow.RegisterIntent(&IntentCheckConflictAndCreateIdenity{})
}

type IntentCheckConflictAndCreateIdenity struct {
	JSONPointer jsonpointer.T          `json:"json_pointer,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Request     *CreateIdentityRequest `json:"request,omitempty"`
}

var _ authflow.Intent = &IntentCheckConflictAndCreateIdenity{}
var _ authflow.Milestone = &IntentCheckConflictAndCreateIdenity{}
var _ MilestoneFlowCreateIdentity = &IntentCheckConflictAndCreateIdenity{}

func (*IntentCheckConflictAndCreateIdenity) Milestone() {}
func (*IntentCheckConflictAndCreateIdenity) MilestoneFlowCreateIdentity(flows authflow.Flows) (MilestoneDoCreateIdentity, authflow.Flows, bool) {
	// If there are no conflicts, then we will have MilestoneDoCreateIdentity directly.
	m, mFlows, ok := authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateIdentity](flows)
	if ok {
		return m, mFlows, true
	}

	// If there are conflicts, then we delegate to MilestoneFlowAccountLinking, which also implements MilestoneFlowCreateIdentity.
	// We cannot find MilestoneFlowCreateIdentity because that would find this intent, resulting in an infinite recursion.
	m1, m1Flows, ok := authflow.FindMilestoneInCurrentFlow[MilestoneFlowAccountLinking](flows)
	if ok {
		return m1.MilestoneFlowCreateIdentity(m1Flows)
	}

	return nil, flows, false
}

func (*IntentCheckConflictAndCreateIdenity) Kind() string {
	return "IntentCheckConflictAndCreateIdenity"
}

func (*IntentCheckConflictAndCreateIdenity) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	switch len(flows.Nearest.Nodes) {
	case 0: // next node is NodeDoCreateIdentity, or account linking intent
		return nil, nil
	}
	return nil, authflow.ErrEOF
}

func (i *IntentCheckConflictAndCreateIdenity) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	switch len(flows.Nearest.Nodes) {
	case 0: // next node is NodeDoCreateIdentity, or account linking intent
		conflicts, err := i.checkConflictByAccountLinkings(ctx, deps, flows)
		if err != nil {
			return nil, err
		}
		spec := i.getIdenitySpec()
		if len(conflicts) == 0 {
			info, err := newIdentityInfo(deps, i.UserID, spec)
			if err != nil {
				return nil, err
			}
			return authflow.NewNodeSimple(&NodeDoCreateIdentity{
				Identity: info,
			}), nil
		} else {
			return authflow.NewSubFlow(&IntentAccountLinking{
				JSONPointer:          i.JSONPointer,
				Conflicts:            conflicts,
				IncomingIdentitySpec: spec,
			}), nil
		}
	}
	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentCheckConflictAndCreateIdenity) checkConflictByAccountLinkings(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows) (conflicts []*AccountLinkingConflict, err error) {
	switch i.Request.Type {
	case model.IdentityTypeOAuth:
		return linkByIncomingOAuthSpec(ctx, deps, flows, i.UserID, i.Request.OAuth, i.JSONPointer)
	case model.IdentityTypeLoginID:
		return linkByIncomingLoginIDSpec(ctx, deps, flows, i.UserID, i.Request.LoginID, i.JSONPointer)
	default:
		// Linking of other types are not supported at the moment
		return nil, nil
	}
}

func (i *IntentCheckConflictAndCreateIdenity) getIdenitySpec() *identity.Spec {
	var spec *identity.Spec
	switch i.Request.Type {
	case model.IdentityTypeLoginID:
		spec = i.Request.LoginID.Spec
	case model.IdentityTypeOAuth:
		spec = i.Request.OAuth.Spec
	default:
		panic(fmt.Errorf("unexpected identity type %v", i.Request.Type))
	}
	return spec
}
