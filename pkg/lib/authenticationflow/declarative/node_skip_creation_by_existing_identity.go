package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeSkipCreationByExistingIdentity{})
}

type NodeSkipCreationByExistingIdentity struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	Identity       *identity.Info                          `json:"identity,omitempty"`
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
}

var _ authflow.NodeSimple = &NodeSkipCreationByExistingIdentity{}
var _ authflow.Milestone = &NodeSkipCreationByExistingIdentity{}
var _ MilestoneIdentificationMethod = &NodeSkipCreationByExistingIdentity{}
var _ authflow.InputReactor = &NodeSkipCreationByExistingIdentity{}

func (*NodeSkipCreationByExistingIdentity) Milestone() {}

func (*NodeSkipCreationByExistingIdentity) Kind() string {
	return "NodeSkipCreationByExistingIdentity"
}

func (n *NodeSkipCreationByExistingIdentity) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return n.Identification
}

func (n *NodeSkipCreationByExistingIdentity) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return nil, nil
}

func (n *NodeSkipCreationByExistingIdentity) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	return authflow.NewNodeSimple(&NodeDoCreateIdentity{
		Identity:   n.Identity,
		SkipCreate: true,
	}), nil
}
