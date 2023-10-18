package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentPromoteIdentityOAuth{})
}

type IntentPromoteIdentityOAuth struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	UserID         string                                  `json:"user_id,omitempty"`
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
	SyntheticInput *InputStepIdentify                      `json:"synthetic_input,omitempty"`
}

var _ authflow.Intent = &IntentPromoteIdentityOAuth{}
var _ authflow.Milestone = &IntentPromoteIdentityOAuth{}
var _ MilestoneIdentificationMethod = &IntentPromoteIdentityOAuth{}

func (*IntentPromoteIdentityOAuth) Kind() string {
	return "IntentPromoteIdentityOAuth"
}

func (*IntentPromoteIdentityOAuth) Milestone() {}
func (i *IntentPromoteIdentityOAuth) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return i.Identification
}

func (i *IntentPromoteIdentityOAuth) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if len(flows.Nearest.Nodes) == 0 {
		oauthOptions := NewIdentificationOptionsOAuth(deps.Config.Identity.OAuth, deps.FeatureConfig.Identity.OAuth.Providers)
		return &InputSchemaTakeOAuthAuthorizationRequest{
			JSONPointer:  i.JSONPointer,
			OAuthOptions: oauthOptions,
		}, nil
	}
	return nil, authflow.ErrEOF
}

func (i *IntentPromoteIdentityOAuth) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	if len(flows.Nearest.Nodes) == 0 {
		var inputOAuth inputTakeOAuthAuthorizationRequest
		if authflow.AsInput(input, &inputOAuth) {
			alias := inputOAuth.GetOAuthAlias()
			redirectURI := inputOAuth.GetOAuthRedirectURI()
			responseMode := inputOAuth.GetOAuthResponseMode()

			syntheticInput := &InputStepIdentify{
				Identification: i.SyntheticInput.Identification,
				Alias:          alias,
				RedirectURI:    redirectURI,
				ResponseMode:   responseMode,
			}

			return authflow.NewNodeSimple(&NodePromoteIdentityOAuth{
				JSONPointer:    i.JSONPointer,
				UserID:         i.UserID,
				SyntheticInput: syntheticInput,
				Alias:          alias,
				RedirectURI:    redirectURI,
				ResponseMode:   responseMode,
			}), nil
		}
	}
	return nil, authflow.ErrIncompatibleInput
}
