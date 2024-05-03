package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaStepIdentify struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
	Options        []IdentificationOption
}

var _ authflow.InputSchema = &InputSchemaStepIdentify{}

func (i *InputSchemaStepIdentify) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaStepIdentify) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaStepIdentify) SchemaBuilder() validation.SchemaBuilder {
	oneOf := []validation.SchemaBuilder{}

	for _, option := range i.Options {
		b := validation.SchemaBuilder{}
		required := []string{"identification"}
		b.Properties().Property("identification", validation.SchemaBuilder{}.Const(option.Identification))

		requireString := func(key string) {
			required = append(required, key)
			b.Properties().Property(key, validation.SchemaBuilder{}.Type(validation.TypeString))
		}

		setRequiredAndAppendOneOf := func() {
			b.Required(required...)
			oneOf = append(oneOf, b)
		}

		switch option.Identification {
		case config.AuthenticationFlowIdentificationIDToken:
			requireString("id_token")
			setRequiredAndAppendOneOf()
		case config.AuthenticationFlowIdentificationEmail:
			requireString("login_id")
			setRequiredAndAppendOneOf()
		case config.AuthenticationFlowIdentificationPhone:
			requireString("login_id")
			setRequiredAndAppendOneOf()
		case config.AuthenticationFlowIdentificationUsername:
			requireString("login_id")
			setRequiredAndAppendOneOf()
		case config.AuthenticationFlowIdentificationOAuth:
			// redirect_uri is required.
			required = append(required, "redirect_uri")
			b.Properties().Property("redirect_uri", validation.SchemaBuilder{}.Type(validation.TypeString).Format("uri"))

			// alias is required.
			required = append(required, "alias")
			b.Properties().Property("alias", validation.SchemaBuilder{}.Type(validation.TypeString).Const(option.Alias))

			// response_mode is optional.
			b.Properties().Property("response_mode", validation.SchemaBuilder{}.
				Type(validation.TypeString).
				Enum(sso.ResponseModeFormPost, sso.ResponseModeQuery))

			setRequiredAndAppendOneOf()
		case config.AuthenticationFlowIdentificationPasskey:
			required = append(required, "assertion_response")
			b.Properties().Property("assertion_response", passkeyAssertionResponseSchemaBuilder)
			setRequiredAndAppendOneOf()
		default:
			break
		}
	}

	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject)

	if len(oneOf) > 0 {
		b.OneOf(oneOf...)
	}

	return b
}

func (i *InputSchemaStepIdentify) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputStepIdentify
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputStepIdentify struct {
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`

	IDToken string `json:"id_token,omitempty"`

	LoginID string `json:"login,omitempty"`

	Alias        string           `json:"alias,omitempty"`
	RedirectURI  string           `json:"redirect_uri,omitempty"`
	ResponseMode sso.ResponseMode `json:"response_mode,omitempty"`
}

var _ authflow.Input = &InputStepIdentify{}
var _ inputTakeIdentificationMethod = &InputStepIdentify{}
var _ inputTakeIDToken = &InputStepIdentify{}
var _ inputTakeLoginID = &InputStepIdentify{}
var _ inputTakeOAuthAuthorizationRequest = &InputStepIdentify{}

func (*InputStepIdentify) Input() {}

func (i *InputStepIdentify) GetIdentificationMethod() config.AuthenticationFlowIdentification {
	return i.Identification
}

func (i *InputStepIdentify) GetIDToken() string {
	return i.IDToken
}

func (i *InputStepIdentify) GetLoginID() string {
	return i.LoginID
}

func (i *InputStepIdentify) GetOAuthAlias() string {
	return i.Alias
}

func (i *InputStepIdentify) GetOAuthRedirectURI() string {
	return i.RedirectURI
}

func (i *InputStepIdentify) GetOAuthResponseMode() sso.ResponseMode {
	return i.ResponseMode
}
