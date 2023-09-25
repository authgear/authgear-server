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
		oauthCandidates := NewIdentificationCandidatesOAuth(deps.Config.Identity.OAuth, deps.FeatureConfig.Identity.OAuth.Providers)
		return &InputSchemaTakeOAuthAuthorizationRequest{
			JSONPointer:     i.JSONPointer,
			OAuthCandidates: oauthCandidates,
		}, nil
	}
	return nil, authflow.ErrEOF
}

func (i *IntentLookupIdentityOAuth) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	if len(flows.Nearest.Nodes) == 0 {
		var inputOAuth inputTakeOAuthAuthorizationRequest
		if authflow.AsInput(input, &inputOAuth) {
			alias := inputOAuth.GetOAuthAlias()
			state := inputOAuth.GetOAuthState()
			redirectURI := inputOAuth.GetOAuthRedirectURI()
			responseMode := inputOAuth.GetOAuthResponseMode()

			syntheticInput := &InputStepIdentify{
				Identification: i.SyntheticInput.Identification,
				Alias:          alias,
				State:          state,
				RedirectURI:    redirectURI,
				ResponseMode:   responseMode,
			}

			return authflow.NewNodeSimple(&NodeLookupIdentityOAuth{
				JSONPointer:    i.JSONPointer,
				SyntheticInput: syntheticInput,
				Alias:          alias,
				State:          state,
				RedirectURI:    redirectURI,
				ResponseMode:   responseMode,
			}), nil
		}
	}
	return nil, authflow.ErrIncompatibleInput
}
