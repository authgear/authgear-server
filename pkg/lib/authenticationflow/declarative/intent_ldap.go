package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterIntent(&IntentLDAP{})
}

type IntentLDAP struct {
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

func (*IntentLDAP) Kind() string {
	return "IntentCreateIdentityLDAP"
}

func (n *IntentLDAP) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return nil, nil
}

func (n *IntentLDAP) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	return nil, nil
}
