package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func init() {
	authflow.RegisterNode(&NodeOAuth{})
}

type NodeOAuth struct {
	JSONPointer  jsonpointer.T `json:"json_pointer,omitempty"`
	NewUserID    string        `json:"new_user_id,omitempty"`
	Alias        string        `json:"alias,omitempty"`
	RedirectURI  string        `json:"redirect_uri,omitempty"`
	ResponseMode string        `json:"response_mode,omitempty"`
}

var _ authflow.NodeSimple = &NodeOAuth{}
var _ authflow.InputReactor = &NodeOAuth{}
var _ authflow.DataOutputer = &NodeOAuth{}

func (*NodeOAuth) Kind() string {
	return "NodeOAuth"
}

func (n *NodeOAuth) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}
	return &InputSchemaTakeOAuthAuthorizationResponse{
		FlowRootObject: flowRootObject,
		JSONPointer:    n.JSONPointer,
	}, nil
}

func (n *NodeOAuth) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var syntheticInputOAuth syntheticInputOAuth
	var inputOAuth inputTakeOAuthAuthorizationResponse
	// The order of the cases is important.
	// We must handle the synthetic input first.
	// It is because if it is synthetic input,
	// then the code has been consumed.
	// Using the code again will definitely fail.
	switch {
	case authflow.AsInput(input, &syntheticInputOAuth):
		spec := syntheticInputOAuth.GetIdentitySpec()
		return n.reactTo(ctx, deps, flows, spec)
	case authflow.AsInput(input, &inputOAuth):
		spec, err := handleOAuthAuthorizationResponse(deps, HandleOAuthAuthorizationResponseOptions{
			Alias:       n.Alias,
			RedirectURI: n.RedirectURI,
		}, inputOAuth)
		if err != nil {
			return nil, err
		}

		return n.reactTo(ctx, deps, flows, spec)
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *NodeOAuth) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	data, err := getOAuthData(ctx, deps, GetOAuthDataOptions{
		RedirectURI:  n.RedirectURI,
		Alias:        n.Alias,
		ResponseMode: n.ResponseMode,
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (n *NodeOAuth) reactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, spec *identity.Spec) (*authflow.Node, error) {
	// signup
	if n.NewUserID != "" {
		return authflow.NewSubFlow(&IntentCheckConflictAndCreateIdenity{
			JSONPointer: n.JSONPointer,
			UserID:      n.NewUserID,
			Request:     NewCreateOAuthIdentityRequest(n.Alias, spec),
		}), nil
	}
	// Else login

	exactMatch, err := findExactOneIdentityInfo(deps, spec)
	if err != nil {
		return nil, err
	}

	newNode, err := NewNodeDoUseIdentityWithUpdate(ctx, deps, flows, exactMatch, spec)
	if err != nil {
		return nil, err
	}

	return authflow.NewNodeSimple(newNode), nil
}
