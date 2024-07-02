package declarative

import (
	"context"
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/facade"
)

func init() {
	authflow.RegisterNode(&IntentUseIdentityPasskey{})
}

type IntentUseIdentityPasskey struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
}

var _ authflow.Intent = &IntentUseIdentityPasskey{}
var _ authflow.Milestone = &IntentUseIdentityPasskey{}
var _ MilestoneIdentificationMethod = &IntentUseIdentityPasskey{}
var _ MilestoneFlowUseIdentity = &IntentUseIdentityPasskey{}
var _ authflow.InputReactor = &IntentUseIdentityPasskey{}

func (*IntentUseIdentityPasskey) Kind() string {
	return "IntentUseIdentityPasskey"
}

func (*IntentUseIdentityPasskey) Milestone() {}
func (n *IntentUseIdentityPasskey) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return n.Identification
}

func (*IntentUseIdentityPasskey) MilestoneFlowUseIdentity(flows authflow.Flows) (MilestoneDoUseIdentity, authflow.Flows, bool) {
	return authflow.FindMilestoneInCurrentFlow[MilestoneDoUseIdentity](flows)
}

func (n *IntentUseIdentityPasskey) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, identified := authflow.FindMilestoneInCurrentFlow[MilestoneDoUseIdentity](flows)
	if identified {
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
	}, nil
}

func (n *IntentUseIdentityPasskey) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputAssertionResponse inputTakePasskeyAssertionResponse
	if authflow.AsInput(input, &inputAssertionResponse) {
		var bpSpecialErr error
		bpRequired, err := IsNodeBotProtectionRequired(ctx, deps, flows, n.JSONPointer)
		if err != nil {
			return nil, err
		}
		if bpRequired {
			token := inputAssertionResponse.(inputTakeBotProtection).GetBotProtectionProviderResponse()
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

		identitySpec := &identity.Spec{
			Type: model.IdentityTypePasskey,
			Passkey: &identity.PasskeySpec{
				AssertionResponse: assertionResponseBytes,
			},
		}

		exactMatch, err := findExactOneIdentityInfo(deps, identitySpec)
		if err != nil {
			return nil, err
		}

		userID := exactMatch.UserID

		authenticatorSpec := &authenticator.Spec{
			Type: model.AuthenticatorTypePasskey,
			Passkey: &authenticator.PasskeySpec{
				AssertionResponse: assertionResponseBytes,
			},
		}

		authenticators, err := deps.Authenticators.List(userID, authenticator.KeepType(model.AuthenticatorTypePasskey))
		if err != nil {
			return nil, err
		}

		authenticatorInfo, verifyResult, err := deps.Authenticators.VerifyOneWithSpec(
			userID,
			model.AuthenticatorTypePasskey,
			authenticators,
			authenticatorSpec,
			&facade.VerifyOptions{
				AuthenticationDetails: facade.NewAuthenticationDetails(
					userID,
					authn.AuthenticationStagePrimary,
					authn.AuthenticationTypePasskey,
				),
			},
		)
		if err != nil {
			return nil, err
		}

		n, err := NewNodeDoUseIdentityPasskey(ctx, flows, &NodeDoUseIdentityPasskey{
			AssertionResponse: assertionResponseBytes,
			Identity:          exactMatch,
			Authenticator:     authenticatorInfo,
			RequireUpdate:     verifyResult.Passkey,
		})
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(n), bpSpecialErr
	}

	return nil, authflow.ErrIncompatibleInput
}
