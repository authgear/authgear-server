package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentOAuth{})
}

type IntentOAuth struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	NewUserID      string                                  `json:"new_user_id,omitempty"`
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
}

var _ authflow.Intent = &IntentOAuth{}
var _ authflow.Milestone = &IntentOAuth{}
var _ MilestoneIdentificationMethod = &IntentOAuth{}
var _ MilestoneFlowCreateIdentity = &IntentOAuth{}
var _ MilestoneFlowUseIdentity = &IntentOAuth{}

func (*IntentOAuth) Kind() string {
	return "IntentOAuth"
}

func (*IntentOAuth) Milestone() {}
func (i *IntentOAuth) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return i.Identification
}

func (*IntentOAuth) MilestoneFlowCreateIdentity(flows authflow.Flows) (MilestoneDoCreateIdentity, authflow.Flows, bool) {
	// Find IntentCheckConflictAndCreateIdenity
	m, mFlows, ok := authflow.FindMilestoneInCurrentFlow[MilestoneFlowCreateIdentity](flows)
	if !ok {
		return nil, mFlows, false
	}

	// Delegate to IntentCheckConflictAndCreateIdenity
	return m.MilestoneFlowCreateIdentity(mFlows)
}

func (*IntentOAuth) MilestoneFlowUseIdentity(flows authflow.Flows) (MilestoneDoUseIdentity, authflow.Flows, bool) {
	return authflow.FindMilestoneInCurrentFlow[MilestoneDoUseIdentity](flows)
}

func (i *IntentOAuth) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
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

		isBotProtectionRequired, err := IsBotProtectionRequired(ctx, deps, flows, i.JSONPointer)
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

func (i *IntentOAuth) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
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

			return authflow.NewNodeSimple(&NodeOAuth{
				JSONPointer:  i.JSONPointer,
				NewUserID:    i.NewUserID,
				Alias:        alias,
				RedirectURI:  redirectURI,
				ResponseMode: responseMode,
			}), bpSpecialErr
		}
	}
	return nil, authflow.ErrIncompatibleInput
}
