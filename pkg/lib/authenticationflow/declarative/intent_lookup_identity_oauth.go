package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentLookupIdentityOAuth{})
}

type IntentLookupIdentityOAuth struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
	SyntheticInput *InputStepIdentify                      `json:"synthetic_input,omitempty"`
}

var _ authflow.Intent = &IntentLookupIdentityOAuth{}
var _ authflow.Milestone = &IntentLookupIdentityOAuth{}
var _ MilestoneIdentificationMethod = &IntentLookupIdentityOAuth{}

func (*IntentLookupIdentityOAuth) Kind() string {
	return "IntentLookupIdentityOAuth"
}

func (*IntentLookupIdentityOAuth) Milestone() {}
func (i *IntentLookupIdentityOAuth) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return i.Identification
}

func (i *IntentLookupIdentityOAuth) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if len(flows.Nearest.Nodes) == 0 {
		flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
		if err != nil {
			return nil, err
		}
		current, err := authflow.FlowObject(flowRootObject, i.JSONPointer)
		if err != nil {
			return nil, err
		}
		var authflowCfg *config.AuthenticationFlowBotProtection = nil
		if currentBranch, ok := current.(config.AuthenticationFlowObjectBotProtectionConfigProvider); ok {
			authflowCfg = currentBranch.GetBotProtectionConfig()
		}

		oauthOptions := NewIdentificationOptionsOAuth(deps.Config.Identity.OAuth, deps.FeatureConfig.Identity.OAuth.Providers, authflowCfg, deps.Config.BotProtection)
		isBotProtectionRequired, err := IsBotProtectionRequired(ctx, flowRootObject, i.JSONPointer)
		if err != nil {
			return nil, err
		}

		return &InputSchemaTakeOAuthAuthorizationRequest{
			FlowRootObject:          flowRootObject,
			JSONPointer:             i.JSONPointer,
			OAuthOptions:            oauthOptions,
			IsBotProtectionRequired: isBotProtectionRequired,
			BotProtectionCfg:        deps.Config.BotProtection,
		}, nil
	}
	return nil, authflow.ErrEOF
}

func (i *IntentLookupIdentityOAuth) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	if len(flows.Nearest.Nodes) == 0 {
		var inputOAuth inputTakeOAuthAuthorizationRequest
		if authflow.AsInput(input, &inputOAuth) {
			var bpSpecialErr error
			bpSpecialErr, err := HandleBotProtection(ctx, deps, flows, i.JSONPointer, input)
			if err != nil {
				return nil, err
			}

			alias := inputOAuth.GetOAuthAlias()
			redirectURI := inputOAuth.GetOAuthRedirectURI()
			responseMode := inputOAuth.GetOAuthResponseMode()

			syntheticInput := &InputStepIdentify{
				Identification: i.SyntheticInput.Identification,
				Alias:          alias,
				RedirectURI:    redirectURI,
				ResponseMode:   responseMode,
			}

			return authflow.NewNodeSimple(&NodeLookupIdentityOAuth{
				JSONPointer:    i.JSONPointer,
				SyntheticInput: syntheticInput,
				Alias:          alias,
				RedirectURI:    redirectURI,
				ResponseMode:   responseMode,
			}), bpSpecialErr
		}
	}
	return nil, authflow.ErrIncompatibleInput
}
