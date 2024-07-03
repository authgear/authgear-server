package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func init() {
	authflow.RegisterIntent(&IntentCreateAuthenticatorPassword{})
}

type IntentCreateAuthenticatorPassword struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	UserID         string                                  `json:"user_id,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.Intent = &IntentCreateAuthenticatorPassword{}
var _ authflow.InputReactor = &IntentCreateAuthenticatorPassword{}
var _ authflow.Milestone = &IntentCreateAuthenticatorPassword{}
var _ MilestoneFlowCreateAuthenticator = &IntentCreateAuthenticatorPassword{}
var _ MilestoneAuthenticationMethod = &IntentCreateAuthenticatorPassword{}

func (*IntentCreateAuthenticatorPassword) Kind() string {
	return "IntentCreateAuthenticatorPassword"
}

func (*IntentCreateAuthenticatorPassword) Milestone() {}
func (*IntentCreateAuthenticatorPassword) MilestoneFlowCreateAuthenticator(flows authflow.Flows) (MilestoneDoCreateAuthenticator, authflow.Flows, bool) {
	return authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateAuthenticator](flows)
}
func (n *IntentCreateAuthenticatorPassword) MilestoneAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return n.Authentication
}

func (n *IntentCreateAuthenticatorPassword) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, created := authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateAuthenticator](flows)
	if created {
		return nil, authflow.ErrEOF
	}
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}

	isBotProtectionRequired, err := IsBotProtectionRequired(ctx, flowRootObject, n.JSONPointer)
	if err != nil {
		return nil, err
	}

	return &InputSchemaTakeNewPassword{
		FlowRootObject:          flowRootObject,
		JSONPointer:             n.JSONPointer,
		IsBotProtectionRequired: isBotProtectionRequired,
	}, nil
}

func (i *IntentCreateAuthenticatorPassword) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputTakeNewPassword inputTakeNewPassword
	if authflow.AsInput(input, &inputTakeNewPassword) {
		var bpSpecialErr error
		bpRequired, err := IsNodeBotProtectionRequired(ctx, deps, flows, i.JSONPointer)
		if err != nil {
			return nil, err
		}
		if bpRequired {
			var inputTakeBotProtection inputTakeBotProtection
			if !authflow.AsInput(input, &inputTakeBotProtection) {
				return nil, authflow.ErrIncompatibleInput
			}

			token := inputTakeBotProtection.GetBotProtectionProviderResponse()
			bpSpecialErr, err = HandleBotProtection(ctx, deps, token)
			if err != nil {
				return nil, err
			}
		}
		authenticatorKind := i.Authentication.AuthenticatorKind()
		newPassword := inputTakeNewPassword.GetNewPassword()
		isDefault, err := authenticatorIsDefault(deps, i.UserID, authenticatorKind)
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
		info, err := deps.Authenticators.NewWithAuthenticatorID(authenticatorID, spec)
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(&NodeDoCreateAuthenticator{
			Authenticator: info,
		}), bpSpecialErr
	}

	return nil, authflow.ErrIncompatibleInput
}
