package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type IntentLookupIdentityLDAP struct {
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
}

var _ authflow.Intent = &IntentLookupIdentityLDAP{}
var _ authflow.Milestone = &IntentLookupIdentityLDAP{}
var _ MilestoneIdentificationMethod = &IntentLookupIdentityLDAP{}

func (*IntentLookupIdentityLDAP) Kind() string {
	return "IntentLookupIdentityLDAP"
}

func (*IntentLookupIdentityLDAP) Milestone() {}

func (i *IntentLookupIdentityLDAP) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return config.AuthenticationFlowIdentificationLDAP
}

func (i *IntentLookupIdentityLDAP) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return nil, nil
}

func (i *IntentLookupIdentityLDAP) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	return nil, nil
}
