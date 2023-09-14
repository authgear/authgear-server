package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeUseIdentityPasskey{})
}

type NodeUseIdentityPasskey struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
	RequestOptions *model.WebAuthnRequestOptions           `json:"request_options,omitempty"`
}

var _ authflow.NodeSimple = &NodeUseIdentityPasskey{}
var _ authflow.Milestone = &NodeUseIdentityPasskey{}
var _ MilestoneIdentificationMethod = &NodeUseIdentityPasskey{}
var _ authflow.InputReactor = &NodeUseIdentityPasskey{}

func (*NodeUseIdentityPasskey) Kind() string {
	return "NodeUseIdentityPasskey"
}

func (*NodeUseIdentityPasskey) Milestone() {}
func (n *NodeUseIdentityPasskey) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return n.Identification
}

func (n *NodeUseIdentityPasskey) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return &InputSchemaTakePasskeyAssertionResponse{
		JSONPointer: n.JSONPointer,
	}, nil
}

func (n *NodeUseIdentityPasskey) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	// FIXME(authflow): handle assertion response.
	return nil, authflow.ErrIncompatibleInput
}
