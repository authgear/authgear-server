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

func (i *IntentLDAP) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if len(flows.Nearest.Nodes) == 0 {
		flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
		if err != nil {
			return nil, err
		}
		return &InputSchemaTakeLDAP{
			FlowRootObject: flowRootObject,
			JSONPointer:    i.JSONPointer,
		}, nil
	}
	return nil, authflow.ErrEOF
}

func (i *IntentLDAP) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputTakeLDAP inputTakeLDAP
	if authflow.AsInput(input, &inputTakeLDAP) {
		entry, err := deps.LDAPClientFactory.Authenticate(
			inputTakeLDAP.GetServer(),
			inputTakeLDAP.GetUserIDAttributeValue(),
			inputTakeLDAP.GetPassword(),
		)
		if err != nil {
			return nil, err
		}

		spec, err := createIdentitySpecFromLDAPEntry(deps, inputTakeLDAP.GetServer(), entry)
		if err != nil {
			return nil, err
		}

		// UserID is the id we assign to new user
		// It is not the user id of an exsiting user
		// Sign up
		if i.UserID != "" {
			return authflow.NewSubFlow(&IntentCheckConflictAndCreateIdenity{
				JSONPointer: i.JSONPointer,
				UserID:      i.UserID,
				Request:     NewCreateLDAPIdentityRequest(spec),
			}), nil
		}

		// login
		exactMatch, err := findExactOneIdentityInfo(deps, spec)
		if err != nil {
			return nil, err
		}

		newNode, err := NewNodeDoUseIdentity(ctx, flows, &NodeDoUseIdentity{
			Identity: exactMatch,
		})
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(newNode), nil
	}
	return nil, nil
}
