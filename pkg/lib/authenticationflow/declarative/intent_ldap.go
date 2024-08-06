package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentLDAP{})
}

type IntentLDAP struct {
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ authflow.Intent = &IntentLDAP{}
var _ authflow.Milestone = &IntentLDAP{}
var _ MilestoneIdentificationMethod = &IntentLDAP{}
var _ MilestoneFlowCreateIdentity = &IntentLDAP{}
var _ MilestoneFlowUseIdentity = &IntentLDAP{}

func (*IntentLDAP) Kind() string {
	return "IntentCreateIdentityLDAP"
}

func (*IntentLDAP) Milestone() {}

func (i *IntentLDAP) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return config.AuthenticationFlowIdentificationLDAP
}

func (*IntentLDAP) MilestoneFlowCreateIdentity(flows authflow.Flows) (MilestoneDoCreateIdentity, authflow.Flows, bool) {
	// Find IntentCheckConflictAndCreateIdenity
	m, mFlows, ok := authflow.FindMilestoneInCurrentFlow[MilestoneFlowCreateIdentity](flows)
	if !ok {
		return nil, mFlows, false
	}

	// Delegate to IntentCheckConflictAndCreateIdenity
	return m.MilestoneFlowCreateIdentity(mFlows)
}

func (*IntentLDAP) MilestoneFlowUseIdentity(flows authflow.Flows) (MilestoneDoUseIdentity, authflow.Flows, bool) {
	return authflow.FindMilestoneInCurrentFlow[MilestoneDoUseIdentity](flows)
}

func (n *IntentLDAP) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return nil, nil
}

func (n *IntentLDAP) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	return nil, nil
}
