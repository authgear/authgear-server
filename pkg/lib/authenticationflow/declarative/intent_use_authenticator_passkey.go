package declarative

import (
	"context"
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/facade"
)

func init() {
	authflow.RegisterIntent(&IntentUseAuthenticatorPasskey{})
}

type IntentUseAuthenticatorPasskey struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	UserID         string                                  `json:"user_id,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.Intent = &IntentUseAuthenticatorPasskey{}
var _ authflow.Milestone = &IntentUseAuthenticatorPasskey{}
var _ MilestoneAuthenticationMethod = &IntentUseAuthenticatorPasskey{}
var _ MilestoneFlowAuthenticate = &IntentUseAuthenticatorPasskey{}
var _ authflow.InputReactor = &IntentUseAuthenticatorPasskey{}

func (*IntentUseAuthenticatorPasskey) Kind() string {
	return "IntentUseAuthenticatorPasskey"
}

func (*IntentUseAuthenticatorPasskey) Milestone() {}
func (n *IntentUseAuthenticatorPasskey) MilestoneAuthenticationMethod() config.AuthenticationFlowAuthentication {
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
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}
	isBotProtectionRequired, err := IsBotProtectionRequired(ctx, flowRootObject, n.JSONPointer)
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

func (n *IntentUseAuthenticatorPasskey) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputAssertionResponse inputTakePasskeyAssertionResponse
	if authflow.AsInput(input, &inputAssertionResponse) {
		var bpSpecialErr error
		bpRequired, err := IsNodeBotProtectionRequired(ctx, deps, flows, n.JSONPointer)
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

		authenticators, err := deps.Authenticators.List(n.UserID, authenticator.KeepType(model.AuthenticatorTypePasskey))
		if err != nil {
			return nil, err
		}

		authenticatorInfo, verifyResult, err := deps.Authenticators.VerifyOneWithSpec(
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
