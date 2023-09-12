package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodeOAuth{})
}

type NodeOAuthData struct {
	AuthorizationURL string `json:"authorization_url,omitempty"`
}

var _ authflow.Data = NodeOAuthData{}

func (NodeOAuthData) Data() {}

type NodeOAuth struct {
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	Alias       string        `json:"alias,omitempty"`
	State       string        `json:"state,omitempty"`
	RedirectURI string        `json:"redirect_uri,omitempty"`
}

var _ authflow.NodeSimple = &NodeOAuth{}
var _ authflow.InputReactor = &NodeOAuth{}
var _ authflow.DataOutputer = &NodeOAuth{}

func (*NodeOAuth) Kind() string {
	return "NodeOAuth"
}

func (n *NodeOAuth) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return &InputSchemaTakeOAuthAuthorizationResponse{
		JSONPointer: n.JSONPointer,
	}, nil
}

func (n *NodeOAuth) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	// FIXME(authflow): handle code or error.
	return nil, authflow.ErrIncompatibleInput
}

func (n *NodeOAuth) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	// FIXME(authflow): compose the authorization URL.
	return NodeOAuthData{}, nil
}
