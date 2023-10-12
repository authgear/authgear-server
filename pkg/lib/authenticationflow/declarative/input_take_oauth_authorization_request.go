package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaTakeOAuthAuthorizationRequest struct {
	JSONPointer  jsonpointer.T
	OAuthOptions []IdentificationOption
}

var _ authflow.InputSchema = &InputSchemaTakeOAuthAuthorizationRequest{}

func (i *InputSchemaTakeOAuthAuthorizationRequest) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakeOAuthAuthorizationRequest) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}.Type(validation.TypeObject)
	b.Required("alias", "redirect_uri")

	b.Properties().Property("redirect_uri", validation.SchemaBuilder{}.Type(validation.TypeString).Format("uri"))
	b.Properties().Property("response_mode", validation.SchemaBuilder{}.
		Type(validation.TypeString).
		Enum(sso.ResponseModeFormPost, sso.ResponseModeQuery))

	var enumValues []interface{}
	for _, c := range i.OAuthOptions {
		enumValues = append(enumValues, c.Alias)

	}
	b.Properties().Property("alias", validation.SchemaBuilder{}.
		Type(validation.TypeString).
		Enum(enumValues...))
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
	Alias        string           `json:"alias"`
	RedirectURI  string           `json:"redirect_uri"`
	ResponseMode sso.ResponseMode `json:"ResponseMode,omitempty"`
}

var _ authflow.Input = &InputTakeOAuthAuthorizationRequest{}
var _ inputTakeOAuthAuthorizationRequest = &InputTakeOAuthAuthorizationRequest{}

func (*InputTakeOAuthAuthorizationRequest) Input() {}

func (i *InputTakeOAuthAuthorizationRequest) GetOAuthAlias() string {
	return i.Alias
}

func (i *InputTakeOAuthAuthorizationRequest) GetOAuthRedirectURI() string {
	return i.RedirectURI
}

func (i *InputTakeOAuthAuthorizationRequest) GetOAuthResponseMode() sso.ResponseMode {
	return i.ResponseMode
}
