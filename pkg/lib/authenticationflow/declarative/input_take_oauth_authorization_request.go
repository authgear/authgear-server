package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaTakeOAuthAuthorizationRequest struct {
	JSONPointer             jsonpointer.T
	FlowRootObject          config.AuthenticationFlowObject
	OAuthOptions            []IdentificationOption
	IsBotProtectionRequired bool
}

var _ authflow.InputSchema = &InputSchemaTakeOAuthAuthorizationRequest{}

func (i *InputSchemaTakeOAuthAuthorizationRequest) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakeOAuthAuthorizationRequest) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaTakeOAuthAuthorizationRequest) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}.Type(validation.TypeObject)
	b.Required("alias", "redirect_uri")

	b.Properties().Property("redirect_uri", validation.SchemaBuilder{}.Type(validation.TypeString).Format("uri"))
	b.Properties().Property("response_mode", validation.SchemaBuilder{}.
		Type(validation.TypeString).
		Enum(oauthrelyingparty.ResponseModeFormPost, oauthrelyingparty.ResponseModeQuery))

	var enumValues []interface{}
	for _, c := range i.OAuthOptions {
		enumValues = append(enumValues, c.Alias)

	}
	b.Properties().Property("alias", validation.SchemaBuilder{}.
		Type(validation.TypeString).
		Enum(enumValues...))
	if i.IsBotProtectionRequired {
		b = AddBotProtectionToExistingSchemaBuilder(b)
	}
	return b
}

func (i *InputSchemaTakeOAuthAuthorizationRequest) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakeOAuthAuthorizationRequest
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputTakeOAuthAuthorizationRequest struct {
	Alias         string                  `json:"alias"`
	RedirectURI   string                  `json:"redirect_uri"`
	ResponseMode  string                  `json:"response_mode,omitempty"`
	BotProtection *InputTakeBotProtection `json:"bot_protection,omitempty"`
}

var _ authflow.Input = &InputTakeOAuthAuthorizationRequest{}
var _ inputTakeOAuthAuthorizationRequest = &InputTakeOAuthAuthorizationRequest{}
var _ inputTakeBotProtection = &InputTakeOAuthAuthorizationRequest{}

func (*InputTakeOAuthAuthorizationRequest) Input() {}

func (i *InputTakeOAuthAuthorizationRequest) GetOAuthAlias() string {
	return i.Alias
}

func (i *InputTakeOAuthAuthorizationRequest) GetOAuthRedirectURI() string {
	return i.RedirectURI
}

func (i *InputTakeOAuthAuthorizationRequest) GetOAuthResponseMode() string {
	return i.ResponseMode
}

func (i *InputTakeOAuthAuthorizationRequest) GetBotProtectionProvider() *InputTakeBotProtection {
	return i.BotProtection
}

func (i *InputTakeOAuthAuthorizationRequest) GetBotProtectionProviderType() config.BotProtectionProviderType {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Type
}

func (i *InputTakeOAuthAuthorizationRequest) GetBotProtectionProviderResponse() string {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Response
}
