package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func init() {
	authflow.RegisterIntent(&IntentCreateAuthenticatorPassword{})
}

type IntentCreateAuthenticatorPassword struct {
	JSONPointer    jsonpointer.T                          `json:"json_pointer,omitempty"`
	UserID         string                                 `json:"user_id,omitempty"`
	Authentication model.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.Intent = &IntentCreateAuthenticatorPassword{}
var _ authflow.InputReactor = &IntentCreateAuthenticatorPassword{}
var _ authflow.Milestone = &IntentCreateAuthenticatorPassword{}
var _ MilestoneFlowCreateAuthenticator = &IntentCreateAuthenticatorPassword{}
var _ MilestoneFlowSelectAuthenticationMethod = &IntentCreateAuthenticatorPassword{}
var _ MilestoneDidSelectAuthenticationMethod = &IntentCreateAuthenticatorPassword{}

func (*IntentCreateAuthenticatorPassword) Kind() string {
	return "IntentCreateAuthenticatorPassword"
}

func (*IntentCreateAuthenticatorPassword) Milestone() {}
func (*IntentCreateAuthenticatorPassword) MilestoneFlowCreateAuthenticator(flows authflow.Flows) (MilestoneDoCreateAuthenticator, authflow.Flows, bool) {
	return authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateAuthenticator](flows)
}
func (n *IntentCreateAuthenticatorPassword) MilestoneFlowSelectAuthenticationMethod(flows authflow.Flows) (MilestoneDidSelectAuthenticationMethod, authflow.Flows, bool) {
	return n, flows, true
}
func (n *IntentCreateAuthenticatorPassword) MilestoneDidSelectAuthenticationMethod() model.AuthenticationFlowAuthentication {
	return n.Authentication
}

func (n *IntentCreateAuthenticatorPassword) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, created := authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateAuthenticator](flows)
	if created {
		return nil, authflow.ErrEOF
	}
	flowRootObject, err := findNearestFlowObjectInFlow(deps, flows, n)
	if err != nil {
		return nil, err
	}

	return &InputSchemaTakeNewPassword{
		FlowRootObject: flowRootObject,
		JSONPointer:    n.JSONPointer,
	}, nil
}

func (i *IntentCreateAuthenticatorPassword) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	var inputTakeNewPassword inputTakeNewPassword
	if authflow.AsInput(input, &inputTakeNewPassword) {
		authenticatorKind := i.Authentication.AuthenticatorKind()
		newPassword := inputTakeNewPassword.GetNewPassword()
		isDefault, err := authenticatorIsDefault(ctx, deps, i.UserID, authenticatorKind)
		if err != nil {
			return nil, err
		}

		spec := &authenticator.Spec{
			UserID:    i.UserID,
			IsDefault: isDefault,
			Kind:      authenticatorKind,
			Type:      model.AuthenticatorTypePassword,
			Password: &authenticator.PasswordSpec{
				PlainPassword: newPassword,
			},
		}

		authenticatorID := uuid.New()
		info, err := deps.Authenticators.NewWithAuthenticatorID(ctx, authenticatorID, spec)
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(&NodeDoCreateAuthenticator{
			Authenticator: info,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}
