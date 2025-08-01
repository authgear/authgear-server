package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterIntent(&IntentLDAP{})
}

type IntentLDAP struct {
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	NewUserID   string        `json:"new_user_id,omitempty"`
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

func (i *IntentLDAP) MilestoneIdentificationMethod() model.AuthenticationFlowIdentification {
	return model.AuthenticationFlowIdentificationLDAP
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
		flowRootObject, err := findNearestFlowObjectInFlow(deps, flows, i)
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

func (i *IntentLDAP) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	var inputTakeLDAP inputTakeLDAP
	if authflow.AsInput(input, &inputTakeLDAP) {
		ldapServerConfig, ok := deps.Config.Identity.LDAP.GetServerConfig(inputTakeLDAP.GetServerName())
		if !ok {
			panic(fmt.Errorf("Unable to find ldap server config with server name %s", inputTakeLDAP.GetServerName()))
		}

		ldapClient := deps.LDAPClientFactory.MakeClient(ldapServerConfig)

		entry, err := ldapClient.AuthenticateUser(
			inputTakeLDAP.GetUsername(),
			inputTakeLDAP.GetPassword(),
		)
		if err != nil {
			return nil, err
		}

		spec, err := deps.LDAP.MakeSpecFromEntry(ctx, ldapServerConfig, inputTakeLDAP.GetUsername(), entry)
		if err != nil {
			return nil, err
		}

		// NewUserID is the id we assign to new user
		// It is not the user id of an existing user
		// Sign up
		if i.NewUserID != "" {
			return authflow.NewSubFlow(&IntentCheckConflictAndCreateIdenity{
				JSONPointer: i.JSONPointer,
				UserID:      i.NewUserID,
				Request:     NewCreateLDAPIdentityRequest(spec),
			}), nil
		}

		// login
		exactMatch, err := findExactOneIdentityInfo(ctx, deps, spec)
		if err != nil {
			return nil, err
		}

		return NewNodeDoUseIdentityWithUpdate(ctx, deps, flows, exactMatch, spec)
	}
	return nil, nil
}
