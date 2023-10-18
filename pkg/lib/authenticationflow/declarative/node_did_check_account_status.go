package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodeDidCheckAccountStatus{})
}

type NodeDidCheckAccountStatusData struct {
	Error *apierrors.APIError `json:"error,omitempty"`
}

var _ authflow.Data = &NodeDidCheckAccountStatusData{}

func (NodeDidCheckAccountStatusData) Data() {}

type NodeDidCheckAccountStatus struct {
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	Error       *apierrors.APIError
}

var _ authflow.NodeSimple = &NodeDidCheckAccountStatus{}
var _ authflow.DataOutputer = &NodeDidCheckAccountStatus{}

func (n *NodeDidCheckAccountStatus) Kind() string {
	return "NodeDidCheckAccountStatus"
}

func (n *NodeDidCheckAccountStatus) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return &InputSchemaCheckAccountStatus{
		JSONPointer: n.JSONPointer,
	}, nil
}

func (n *NodeDidCheckAccountStatus) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, _ authflow.Input) (*authflow.Node, error) {
	// But nothing is compatible.
	return nil, authflow.ErrIncompatibleInput
}

func (n *NodeDidCheckAccountStatus) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return NodeDidCheckAccountStatusData{
		Error: n.Error,
	}, nil
}
