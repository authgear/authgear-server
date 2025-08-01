package declarative

import (
	"context"
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/facade"
)

func init() {
	authflow.RegisterIntent(&IntentUseAuthenticatorPasskey{})
}

type IntentUseAuthenticatorPasskey struct {
	JSONPointer    jsonpointer.T                          `json:"json_pointer,omitempty"`
	UserID         string                                 `json:"user_id,omitempty"`
	Authentication model.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.Intent = &IntentUseAuthenticatorPasskey{}
var _ authflow.Milestone = &IntentUseAuthenticatorPasskey{}
var _ MilestoneFlowSelectAuthenticationMethod = &IntentUseAuthenticatorPasskey{}
var _ MilestoneDidSelectAuthenticationMethod = &IntentUseAuthenticatorPasskey{}
var _ MilestoneFlowAuthenticate = &IntentUseAuthenticatorPasskey{}
var _ authflow.InputReactor = &IntentUseAuthenticatorPasskey{}

func (*IntentUseAuthenticatorPasskey) Kind() string {
	return "IntentUseAuthenticatorPasskey"
}

func (*IntentUseAuthenticatorPasskey) Milestone() {}
func (n *IntentUseAuthenticatorPasskey) MilestoneFlowSelectAuthenticationMethod(flows authflow.Flows) (MilestoneDidSelectAuthenticationMethod, authflow.Flows, bool) {
	return n, flows, true
}
func (n *IntentUseAuthenticatorPasskey) MilestoneDidSelectAuthenticationMethod() model.AuthenticationFlowAuthentication {
	return n.Authentication
}

func (*IntentUseAuthenticatorPasskey) MilestoneFlowAuthenticate(flows authflow.Flows) (MilestoneDidAuthenticate, authflow.Flows, bool) {
	return authflow.FindMilestoneInCurrentFlow[MilestoneDidAuthenticate](flows)
}

func (n *IntentUseAuthenticatorPasskey) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, authenticated := authflow.FindMilestoneInCurrentFlow[MilestoneDidAuthenticate](flows)
	if authenticated {
		return nil, authflow.ErrEOF
	}
	flowRootObject, err := findNearestFlowObjectInFlow(deps, flows, n)
	if err != nil {
		return nil, err
	}
	isBotProtectionRequired, err := IsBotProtectionRequired(ctx, deps, flows, n.JSONPointer, n)
	if err != nil {
		return nil, err
	}
	return &InputSchemaTakePasskeyAssertionResponse{
		FlowRootObject:          flowRootObject,
		JSONPointer:             n.JSONPointer,
		IsBotProtectionRequired: isBotProtectionRequired,
		BotProtectionCfg:        deps.Config.BotProtection,
	}, nil
}

func (n *IntentUseAuthenticatorPasskey) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	var inputAssertionResponse inputTakePasskeyAssertionResponse
	if authflow.AsInput(input, &inputAssertionResponse) {
		var bpSpecialErr error
		bpSpecialErr, err := HandleBotProtection(ctx, deps, flows, n.JSONPointer, input, n)
		if err != nil {
			return nil, err
		}
		assertionResponse := inputAssertionResponse.GetAssertionResponse()
		assertionResponseBytes, err := json.Marshal(assertionResponse)
		if err != nil {
			return nil, err
		}

		authenticatorSpec := &authenticator.Spec{
			Type: model.AuthenticatorTypePasskey,
			Passkey: &authenticator.PasskeySpec{
				AssertionResponse: assertionResponseBytes,
			},
		}

		authenticators, err := deps.Authenticators.List(ctx, n.UserID, authenticator.KeepType(model.AuthenticatorTypePasskey))
		if err != nil {
			return nil, err
		}

		authenticatorInfo, verifyResult, err := deps.Authenticators.VerifyOneWithSpec(ctx,
			n.UserID,
			model.AuthenticatorTypePasskey,
			authenticators,
			authenticatorSpec,
			&facade.VerifyOptions{
				AuthenticationDetails: facade.NewAuthenticationDetails(
					n.UserID,
					authn.AuthenticationStagePrimary,
					authn.AuthenticationTypePasskey,
				),
			},
		)
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(&NodeDoUseAuthenticatorPasskey{
			AssertionResponse: assertionResponseBytes,
			Authenticator:     authenticatorInfo,
			RequireUpdate:     verifyResult.Passkey,
		}), bpSpecialErr
	}

	return nil, authflow.ErrIncompatibleInput
}
