package declarative

import (
	"context"
	"errors"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
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

func (i *IntentLookupIdentityLDAP) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	if len(flows.Nearest.Nodes) == 0 {
		flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
		if err != nil {
			return nil, err
		}
		current, err := authflow.FlowObject(flowRootObject, i.JSONPointer)
		if err != nil {
			return nil, err
		}

		oneOf := i.oneOf(current)

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

			spec, err := createIdentitySpecFromLDAPEntry(deps, ldapServerConfig, entry)
			if err != nil {
				return nil, err
			}

			syntheticInput := &SyntheticInputLDAP{
				ServerName: inputTakeLDAP.GetServerName(),
				Username:   inputTakeLDAP.GetUsername(),
				Password:   inputTakeLDAP.GetPassword(),
			}

			_, err = findExactOneIdentityInfo(deps, spec)
			if err != nil {
				if apierrors.IsKind(err, api.UserNotFound) {
					// signup
					return nil, errors.Join(&authflow.ErrorSwitchFlow{
						FlowReference: authflow.FlowReference{
							Type: authflow.FlowTypeSignup,
							Name: oneOf.SignupFlow,
						},
						SyntheticInput: syntheticInput,
					})
				}
				// general error
				return nil, err
			}

			// login
			return nil, errors.Join(&authflow.ErrorSwitchFlow{
				FlowReference: authflow.FlowReference{
					Type: authflow.FlowTypeLogin,
					Name: oneOf.LoginFlow,
				},
				SyntheticInput: syntheticInput,
			})
		}
	}
	return nil, authflow.ErrIncompatibleInput
}

func (*IntentLookupIdentityLDAP) oneOf(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupLoginFlowOneOf {
	oneOf, ok := o.(*config.AuthenticationFlowSignupLoginFlowOneOf)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return oneOf
}
