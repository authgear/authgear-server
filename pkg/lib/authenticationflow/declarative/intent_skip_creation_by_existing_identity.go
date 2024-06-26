package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentSkipCreationByExistingIdentity{})
}

type IntentSkipCreationByExistingIdentity struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	Identity       *identity.Info                          `json:"identity,omitempty"`
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
}

var _ authflow.Intent = &IntentSkipCreationByExistingIdentity{}
var _ authflow.Milestone = &IntentSkipCreationByExistingIdentity{}
var _ MilestoneIdentificationMethod = &IntentSkipCreationByExistingIdentity{}
var _ MilestoneFlowCreateIdentity = &IntentSkipCreationByExistingIdentity{}
var _ MilestoneDoCreateIdentity = &IntentSkipCreationByExistingIdentity{}
var _ authflow.InputReactor = &IntentSkipCreationByExistingIdentity{}

func (*IntentSkipCreationByExistingIdentity) Milestone() {}

func (*IntentSkipCreationByExistingIdentity) Kind() string {
	return "IntentSkipCreationByExistingIdentity"
}

func (n *IntentSkipCreationByExistingIdentity) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return n.Identification
}
func (n *IntentSkipCreationByExistingIdentity) MilestoneFlowCreateIdentity(flows authflow.Flows) (MilestoneDoCreateIdentity, authflow.Flows, bool) {
	return n, flows, true
}
func (n *IntentSkipCreationByExistingIdentity) MilestoneDoCreateIdentity() *identity.Info {
	return n.Identity
}
func (n *IntentSkipCreationByExistingIdentity) MilestoneDoCreateIdentitySkipCreate() {
	// Already skipping
}
func (n *IntentSkipCreationByExistingIdentity) MilestoneDoCreateIdentityUpdate(newInfo *identity.Info) {
	n.Identity = newInfo
}

func (n *IntentSkipCreationByExistingIdentity) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, identified := authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateIdentity](flows)
	if identified {
		return nil, authflow.ErrEOF
	}
	return nil, nil
}

func (n *IntentSkipCreationByExistingIdentity) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	return authflow.NewNodeSimple(&NodeDoCreateIdentity{
		Identity:   n.Identity,
		SkipCreate: true,
	}), nil
}
